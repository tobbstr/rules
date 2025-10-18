package rules

import (
	"fmt"
	"testing"
)

// Test types for cross-type rule composition
type User struct {
	ID       string
	Name     string
	Role     string
	IsActive bool
}

type Order struct {
	ID        string
	Amount    float64
	ItemCount int
	Status    string
}

type OrderRequest struct {
	Order Order
	User  User
}

func TestMap(t *testing.T) {
	t.Parallel()

	// Create a rule that operates on User
	userRule := New(
		"user is admin",
		func(user User) (bool, error) {
			return user.Role == "admin", nil
		},
	)

	// Map it to operate on OrderRequest
	mappedRule := Map(
		"request from admin",
		userRule,
		func(req OrderRequest) User {
			return req.User
		},
	)

	tests := []struct {
		name    string
		request OrderRequest
		want    bool
	}{
		{
			name: "admin user",
			request: OrderRequest{
				User: User{Role: "admin"},
			},
			want: true,
		},
		{
			name: "non-admin user",
			request: OrderRequest{
				User: User{Role: "customer"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := mappedRule.Evaluate(tt.request)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapWithNilRule(t *testing.T) {
	t.Parallel()

	mappedRule := Map[OrderRequest, User](
		"nil rule",
		nil,
		func(req OrderRequest) User {
			return req.User
		},
	)

	_, err := mappedRule.Evaluate(OrderRequest{})
	if err == nil {
		t.Error("Expected error for nil rule")
	}
}

func TestMapNameAndDescription(t *testing.T) {
	t.Parallel()

	userRule := New(
		"user rule",
		func(user User) (bool, error) {
			return true, nil
		},
	)

	mappedRule := Map(
		"mapped rule",
		userRule,
		func(req OrderRequest) User {
			return req.User
		},
	)

	if mappedRule.Name() != "mapped rule" {
		t.Errorf(
			"Name() = %q, want %q",
			mappedRule.Name(),
			"mapped rule",
		)
	}

	// Note: Description() method was removed from Rule interface.
	// Descriptions are now stored in the registry as metadata.
}

func TestCombine(t *testing.T) {
	t.Parallel()

	// Rule for User
	userRule := New(
		"user is active admin",
		func(user User) (bool, error) {
			return user.Role == "admin" && user.IsActive, nil
		},
	)

	// Rule for Order
	orderRule := New(
		"order is valid",
		func(order Order) (bool, error) {
			return order.Amount > 0 && order.ItemCount > 0, nil
		},
	)

	// Combine them
	combinedRule := Combine(
		"valid order request",
		userRule,
		func(req OrderRequest) User { return req.User },
		orderRule,
		func(req OrderRequest) Order { return req.Order },
	)

	tests := []struct {
		name    string
		request OrderRequest
		want    bool
	}{
		{
			name: "valid request",
			request: OrderRequest{
				User: User{
					Role:     "admin",
					IsActive: true,
				},
				Order: Order{
					Amount:    100.0,
					ItemCount: 5,
				},
			},
			want: true,
		},
		{
			name: "invalid user",
			request: OrderRequest{
				User: User{
					Role:     "customer",
					IsActive: true,
				},
				Order: Order{
					Amount:    100.0,
					ItemCount: 5,
				},
			},
			want: false,
		},
		{
			name: "invalid order",
			request: OrderRequest{
				User: User{
					Role:     "admin",
					IsActive: true,
				},
				Order: Order{
					Amount:    0,
					ItemCount: 0,
				},
			},
			want: false,
		},
		{
			name: "both invalid",
			request: OrderRequest{
				User: User{
					Role:     "customer",
					IsActive: false,
				},
				Order: Order{
					Amount:    0,
					ItemCount: 0,
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := combinedRule.Evaluate(tt.request)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCombine3(t *testing.T) {
	t.Parallel()

	type ShippingInfo struct {
		Country string
		Express bool
	}

	type CompleteRequest struct {
		User     User
		Order    Order
		Shipping ShippingInfo
	}

	userRule := New(
		"user is active",
		func(user User) (bool, error) {
			return user.IsActive, nil
		},
	)

	orderRule := New(
		"order has items",
		func(order Order) (bool, error) {
			return order.ItemCount > 0, nil
		},
	)

	shippingRule := New(
		"valid shipping country",
		func(shipping ShippingInfo) (bool, error) {
			validCountries := map[string]bool{
				"US": true,
				"CA": true,
				"UK": true,
			}
			return validCountries[shipping.Country], nil
		},
	)

	combinedRule := Combine3(
		"complete request validation",
		userRule,
		func(req CompleteRequest) User { return req.User },
		orderRule,
		func(req CompleteRequest) Order { return req.Order },
		shippingRule,
		func(req CompleteRequest) ShippingInfo { return req.Shipping },
	)

	tests := []struct {
		name    string
		request CompleteRequest
		want    bool
	}{
		{
			name: "all valid",
			request: CompleteRequest{
				User:     User{IsActive: true},
				Order:    Order{ItemCount: 3},
				Shipping: ShippingInfo{Country: "US"},
			},
			want: true,
		},
		{
			name: "invalid shipping",
			request: CompleteRequest{
				User:     User{IsActive: true},
				Order:    Order{ItemCount: 3},
				Shipping: ShippingInfo{Country: "XX"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := combinedRule.Evaluate(tt.request)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCombineMany(t *testing.T) {
	t.Parallel()

	userActiveRule := New(
		"user is active",
		func(user User) (bool, error) {
			return user.IsActive, nil
		},
	)

	userAdminRule := New(
		"user is admin",
		func(user User) (bool, error) {
			return user.Role == "admin", nil
		},
	)

	orderValidRule := New(
		"order is valid",
		func(order Order) (bool, error) {
			return order.Amount > 0, nil
		},
	)

	orderPendingRule := New(
		"order is pending",
		func(order Order) (bool, error) {
			return order.Status == "pending", nil
		},
	)

	// Combine using CombineMany
	combinedRule := CombineMany(
		"all requirements",
		Map(
			"user active check",
			userActiveRule,
			func(req OrderRequest) User { return req.User },
		),
		Map(
			"user admin check",
			userAdminRule,
			func(req OrderRequest) User { return req.User },
		),
		Map(
			"order valid check",
			orderValidRule,
			func(req OrderRequest) Order { return req.Order },
		),
		Map(
			"order pending check",
			orderPendingRule,
			func(req OrderRequest) Order { return req.Order },
		),
	)

	request := OrderRequest{
		User: User{
			IsActive: true,
			Role:     "admin",
		},
		Order: Order{
			Amount: 100.0,
			Status: "pending",
		},
	}

	satisfied, err := combinedRule.Evaluate(request)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !satisfied {
		t.Error("Expected combined rule to be satisfied")
	}
}

// TestMapWithContextCancellation removed - context no longer used

// Example demonstrating cross-type rule composition
func ExampleMap() {
	type User struct {
		Role string
	}

	type Request struct {
		User User
	}

	// Rule that operates on User
	userRule := New(
		"user is admin",
		func(user User) (bool, error) {
			return user.Role == "admin", nil
		},
	)

	// Map it to operate on Request
	requestRule := Map(
		"request from admin",
		userRule,
		func(req Request) User {
			return req.User
		},
	)

	// Evaluate
	request := Request{User: User{Role: "admin"}}
	satisfied, _ := requestRule.Evaluate(request)

	fmt.Printf("Request from admin: %v\n", satisfied)
	// Output: Request from admin: true
}

// Example demonstrating combining rules from different types
func ExampleCombine() {
	type User struct {
		IsActive bool
	}

	type Order struct {
		Amount float64
	}

	type OrderRequest struct {
		User  User
		Order Order
	}

	userRule := New(
		"user is active",
		func(user User) (bool, error) {
			return user.IsActive, nil
		},
	)

	orderRule := New(
		"order has amount",
		func(order Order) (bool, error) {
			return order.Amount > 0, nil
		},
	)

	// Combine rules from different types
	combinedRule := Combine(
		"valid order request",
		userRule,
		func(req OrderRequest) User { return req.User },
		orderRule,
		func(req OrderRequest) Order { return req.Order },
	)

	request := OrderRequest{
		User:  User{IsActive: true},
		Order: Order{Amount: 100.0},
	}

	satisfied, _ := combinedRule.Evaluate(request)

	fmt.Printf("Order request valid: %v\n", satisfied)
	// Output: Order request valid: true
}
