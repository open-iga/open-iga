package domain

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	// WorkflowOnboarding is the name of the predefined onboarding workflow.
	WorkflowOnboarding = "onboarding"
	// A function is a connector operation, named "<kind>.<action>", so the application can run any step through the connector service.
	FunctionCreateAccount    = string(AccountActionKind) + ".create"
	FunctionGrantEntitlement = string(EntitlementKind) + ".grant"
	// The arguments of the functions above: the managed system, the resource of its connector, the tenant and the entitlement.
	ArgSystem        = "system"
	ArgResource      = "resource"
	ArgTenantId      = "tenantId"
	ArgEntitlementId = "entitlementId"
)

// OnboardingEntitlement is an entitlement the admin picked from the ones discovered for the tenant.
type OnboardingEntitlement struct {
	Id uuid.UUID `json:"id"`
	// Gate holds this grant until someone continues it manually.
	Gate bool `json:"gate,omitempty"`
}

// OnboardingStep is one step of onboarding a user (identity) using a managed system: it creates the user's account
// on the tenant, then grants the picked entitlements.
type OnboardingStep struct {
	System       string                  `json:"system"`
	Resource     string                  `json:"resource"`
	TenantId     uuid.UUID               `json:"tenantId"`
	Entitlements []OnboardingEntitlement `json:"entitlements,omitempty"`
}

func (s OnboardingStep) Validate() error {
	if s.System == "" || s.Resource == "" || s.TenantId == uuid.Nil {
		return fmt.Errorf("%w: onboarding step needs a system, a resource and a tenant", ErrInvalidWorkflow)
	}
	return nil
}

// createAccountStepName is also what identifies the step to the UI.
func (s OnboardingStep) createAccountStepName() string {
	return "create-account:" + s.System
}

// WorkflowSteps are the generic steps that run this onboarding step: the account create, then one grant per entitlement.
// The grants wait for the account only, so they run in parallel; a gate or a failure holds that grant alone.
func (s OnboardingStep) WorkflowSteps() []WorkflowStep {
	args := map[string]string{ArgSystem: s.System, ArgResource: s.Resource, ArgTenantId: s.TenantId.String()}
	createName := s.createAccountStepName()
	steps := []WorkflowStep{{Name: createName, Function: FunctionCreateAccount, Args: args, After: nil, Gate: false}}

	for _, entitlement := range s.Entitlements {
		grantArgs := map[string]string{
			ArgSystem: s.System, ArgResource: s.Resource, ArgTenantId: s.TenantId.String(), ArgEntitlementId: entitlement.Id.String(),
		}
		steps = append(steps, WorkflowStep{
			Name:     fmt.Sprintf("grant:%s:%s", s.System, entitlement.Id),
			Function: FunctionGrantEntitlement,
			Args:     grantArgs,
			After:    []string{createName},
			Gate:     entitlement.Gate,
		})
	}
	return steps
}

// NewOnboardingDefinition is the onboarding workflow: on each managed system, an account and its entitlements, one lane per system.
// Lanes run in parallel.
// It holds only what the admin configured; who is onboarded is the input of each run.
func NewOnboardingDefinition(steps []OnboardingStep) (WorkflowDefinition, error) {
	var workflowSteps []WorkflowStep
	for _, step := range steps {
		if err := step.Validate(); err != nil {
			return WorkflowDefinition{}, err
		}
		workflowSteps = append(workflowSteps, step.WorkflowSteps()...)
	}

	definition := WorkflowDefinition{Name: WorkflowOnboarding, Steps: workflowSteps}
	if err := definition.Validate(); err != nil {
		return WorkflowDefinition{}, err
	}
	return definition, nil
}
