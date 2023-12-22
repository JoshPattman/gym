package gym

import (
	"math"
	"math/rand"

	"github.com/gopxl/pixel"
	"github.com/gopxl/pixel/imdraw"
)

var _ Env = &CartPoleEnv{}

// CartPoleSettings contains all the settings for the cartpole environment.
type CartPoleSettings struct {
	// Acceleration of the cart.
	Acceleration float64
	// Max velocity of the cart.
	MaxVelocity float64
	// Max rotational velocity of the pole.
	MaxRotationalVelocity float64
	// Acceleration due to gravity.
	GravityAcceleration float64
	// Torque multiplier of torque applied by cart onto pole.
	TorqueMultiplier float64

	// The delta time between steps.
	TimeStep float64
	// The max initial angle of the pole upon reset.
	MaxInitialAngle float64
	// The max initial offset of the cart upon reset. Should be no more than 1.
	MaxInitialOffset float64
	// The angle at which the pole is considered to have failed.
	FailAngle float64

	// The reward for being centered. This linearly falls off the further we are from the center.
	CenteredPerStepReward float64
	// The reward for going out of bounds. This should be negative.
	OutOfBoundsReward float64
	// The reward for the pole falling over. This should be negative.
	PoleFallReward float64
}

var DefaultCartPoleSettings = CartPoleSettings{
	Acceleration:          0.5,
	MaxVelocity:           1.5,
	MaxRotationalVelocity: math.Pi * 3,
	GravityAcceleration:   9.8,
	TorqueMultiplier:      4.0,

	TimeStep:         1.0 / 60.0,
	MaxInitialAngle:  math.Pi / 8,
	MaxInitialOffset: 0.8,
	FailAngle:        math.Pi / 2,

	CenteredPerStepReward: 1.0,
	OutOfBoundsReward:     -1.0,
	PoleFallReward:        -5.0,
}

type CartPoleEnv struct {
	// The position of the box. It is normalized to be between -1 and 1.
	BoxPosition float64
	// The velocity of the box.
	BoxVelocity float64
	// The rotation of the pole in radians.
	PoleRotation float64
	// The rotational velocity of the pole.
	PoleRotationalVelocity float64
	// The settings for the cartpole environment.
	Settings CartPoleSettings

	drawer *imdraw.IMDraw
}

// NewCartPoleEnv creates a new cartpole environment with the given settings.
func NewCartPoleEnv(settings CartPoleSettings) *CartPoleEnv {
	return &CartPoleEnv{
		Settings: settings,
		drawer:   imdraw.New(nil),
	}
}

// Step performs a step in the environment.
// The action is [left_right_move(-1 to 1): the acceleration to apply to the cart left/right]
// The observation is [cart_position(-1 to 1): the position of the cart, cart_velocity(-1 to 1): the velocity of the cart, pole_angle(-1 to 1): the angle of the pole, pole_angular_velocity(-1 to 1): the angular velocity of the pole]
func (e *CartPoleEnv) Step(action []float64) StepData {
	validateAction(action, e.ActionLength())

	forceAction := action[0]

	// Update box velocity and position.
	e.BoxVelocity += forceAction * e.Settings.Acceleration * e.Settings.TimeStep
	if e.BoxVelocity > e.Settings.MaxVelocity {
		e.BoxVelocity = e.Settings.MaxVelocity
	} else if e.BoxVelocity < -e.Settings.MaxVelocity {
		e.BoxVelocity = -e.Settings.MaxVelocity
	}
	e.BoxPosition += e.BoxVelocity * e.Settings.TimeStep

	// Update pole rotational velocity and rotation.
	poleGravityAcceleration := e.Settings.GravityAcceleration * math.Sin(e.PoleRotation)
	poleTorque := forceAction * math.Cos(e.PoleRotation) * e.Settings.TorqueMultiplier

	e.PoleRotationalVelocity += (poleGravityAcceleration + poleTorque) * e.Settings.TimeStep
	if e.PoleRotationalVelocity > e.Settings.MaxRotationalVelocity {
		e.PoleRotationalVelocity = e.Settings.MaxRotationalVelocity
	} else if e.PoleRotationalVelocity < -e.Settings.MaxRotationalVelocity {
		e.PoleRotationalVelocity = -e.Settings.MaxRotationalVelocity
	}
	e.PoleRotation += e.PoleRotationalVelocity * e.Settings.TimeStep

	// Check if we failed, and find the reward
	failed := false
	reward := e.Settings.CenteredPerStepReward*1 - math.Abs(e.BoxPosition) // Reward falls off the further we are from the center.
	if e.BoxPosition > 1.0 || e.BoxPosition < -1.0 {
		failed = true
		reward = e.Settings.OutOfBoundsReward
	} else if e.PoleRotation > e.Settings.FailAngle || e.PoleRotation < -e.Settings.FailAngle {
		failed = true
		reward = e.Settings.PoleFallReward
	}

	// Return the step data.
	return StepData{
		Observation: e.getObservation(),
		Reward:      reward,
		Terminated:  failed,
		Info:        e.getInfo(),
	}
}

func (e *CartPoleEnv) getObservation() []float64 {
	return clampAll(
		e.BoxPosition,
		e.BoxVelocity/e.Settings.MaxVelocity,
		e.PoleRotation/180.0,
		e.PoleRotationalVelocity/e.Settings.MaxRotationalVelocity,
	)
}

func (e *CartPoleEnv) getInfo() map[string]interface{} {
	return map[string]interface{}{}
}

// Reset resets the environment.
func (e *CartPoleEnv) Reset() ResetData {
	e.BoxPosition = (rand.Float64()*2 - 1) * e.Settings.MaxInitialOffset
	e.BoxVelocity = 0.0
	e.PoleRotation = (rand.Float64()*2 - 1) * e.Settings.MaxInitialAngle
	e.PoleRotationalVelocity = 0.0
	return ResetData{
		Observation: e.getObservation(),
		Info:        e.getInfo(),
	}
}

func (e *CartPoleEnv) Name() string {
	return "CartPole"
}

func (e *CartPoleEnv) RenderSize() (float64, float64) {
	return 1200, 800
}

func (e *CartPoleEnv) Render(target pixel.Target) {
	rsx, rsy := e.RenderSize()
	axisYPos := rsy / 3

	// Draw a rectangle over the whole window.
	e.drawer.Clear()
	e.drawer.Color = pixel.RGB(1, 1, 1)
	e.drawer.Push(pixel.V(0, 0))
	e.drawer.Push(pixel.V(rsx, rsy))
	e.drawer.Rectangle(0)

	// Draw the axis.
	e.drawer.Color = pixel.RGB(0, 0, 0)
	e.drawer.Push(pixel.V(0, axisYPos))
	e.drawer.Push(pixel.V(rsx, axisYPos))
	e.drawer.Line(2)

	// Draw the cart.
	cartXPos := (rsx / 2) + (e.BoxPosition * rsx / 2)
	cartXSize := 50.0
	cartYSize := 35.0
	e.drawer.Color = pixel.RGB(0, 0, 0)
	e.drawer.Push(pixel.V(cartXPos-(cartXSize/2), axisYPos-(cartYSize/2)))
	e.drawer.Push(pixel.V(cartXPos+(cartXSize/2), axisYPos+(cartYSize/2)))
	e.drawer.Rectangle(0)

	// Draw the pole.
	poleBottomPos := pixel.V(cartXPos, axisYPos)
	poleTopPos := poleBottomPos.Add(pixel.V(0, 200.0).Rotated(e.PoleRotation))
	e.drawer.Color = pixel.RGB(0.976, 0.682, 0.357)
	e.drawer.Push(poleBottomPos)
	e.drawer.Push(poleTopPos)
	e.drawer.Line(10)

	e.drawer.Color = pixel.RGB(0.243, 0.396, 0.663)
	e.drawer.Push(poleBottomPos)
	e.drawer.Circle(4, 0)

	e.drawer.Draw(target)
}

// ActionLength returns the length of the action vector.
func (e *CartPoleEnv) ActionLength() int {
	return 1
}

// ObservationLength returns the length of the observation vector.
func (e *CartPoleEnv) ObservationLength() int {
	return len(e.getObservation())
}

func (e *CartPoleEnv) NumCategoricalActions() int {
	return 3
}

// ConvertCategoricalAction converts a categorical action to a continuous action. CAction 0 returns [0], CAction 1 returns [1], CAction 2 returns [-1].
func (e *CartPoleEnv) ConvertCategoricalAction(action int) []float64 {
	switch action {
	case 0:
		return []float64{0.0}
	case 1:
		return []float64{1.0}
	case 2:
		return []float64{-1.0}
	default:
		panic("Invalid action")
	}
}
