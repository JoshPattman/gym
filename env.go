package gym

import "github.com/gopxl/pixel"

// StepData is the data returned by the Step function.
type StepData struct {
	// Observation is a list of floats between -1 and 1.
	// It represents what the agent can sense in the environmnet.
	Observation []float64
	// Reward is the reward for the step we just took.
	Reward float64
	// Terminated is true if the episode is over.
	Terminated bool
	// Info is a map of extra information. This is environment specific and really just for debugging.
	Info map[string]interface{}
}

// ResetData is the data returned by the Reset function.
type ResetData struct {
	// Observation is a list of floats between -1 and 1.
	// It represents what the agent can sense in the environmnet.
	Observation []float64
	// Info is a map of extra information. This is environment specific and really just for debugging.
	Info map[string]interface{}
}

type Env interface {
	// Name gets the name of the environment. E.g. 'CartPole'
	Name() string

	// Render renders the environment to the target.
	// It should remember to also clear the background.
	Render(target pixel.Target)
	// RenderSize specifies the dimensions that the environment should be rendered at.
	// This sets the window size.
	RenderSize() (float64, float64)

	// Step takes an action and steps the environment one timestep forwards.
	Step(action []float64) StepData
	// Reset resets the environment.
	Reset() ResetData

	// ConvertCategoricalAction converts a categorical action to a one-hot vector.
	// For example, action '2' might become {0.5, 0.25, -1}.
	ConvertCategoricalAction(int) []float64
	// NumCategoricalActions gets the number of categorical actions that the environment supports.
	NumCategoricalActions() int
	// ActionLength gets the length of the action vector.
	ActionLength() int
	// ObservationLength gets the length of the observation vector.
	ObservationLength() int
}
