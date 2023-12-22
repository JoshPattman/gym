package gym

import (
	"math"
	"math/rand"

	"github.com/gopxl/pixel"
	"github.com/gopxl/pixel/imdraw"
)

var _ Env = &BallPushEnv{}

var DefaultBallPushSettings = &BallPushSettings{
	BallRadius:          2,
	AgentRadius:         1,
	AgentAcceleration:   20,
	AgentDrag:           1.5,
	BallDrag:            0.75,
	BoundaryRadius:      50,
	MoveToBallReward:    0.5,
	TouchBallReward:     1,
	MoveToCenterReward:  1,
	PlaceInCenterReward: 2,
	Scale:               10,
	TargetRadius:        3,
	DeltaTime:           1.0 / 60.0,
}

type BallPushSettings struct {
	BallRadius          float64
	AgentRadius         float64
	AgentAcceleration   float64
	AgentDrag           float64
	BallDrag            float64
	BoundaryRadius      float64
	MoveToBallReward    float64
	TouchBallReward     float64
	MoveToCenterReward  float64
	PlaceInCenterReward float64
	TargetRadius        float64
	Scale               float64
	DeltaTime           float64
}

func (b *BallPushEnv) BallInCenter() bool {
	return b.Ball.Position().Len() < b.Settings.TargetRadius-b.Settings.BallRadius
}

func (b *BallPushSettings) AgentMaxSpeed() float64 {
	return b.AgentAcceleration / b.AgentDrag
}

type BallPushEnv struct {
	Agent           *VerletParticle
	Ball            *VerletParticle
	HasTouchedBall  bool
	HasCenteredBall bool
	Settings        *BallPushSettings

	imd *imdraw.IMDraw
}

func NewBallPushEnv(settings *BallPushSettings) *BallPushEnv {
	e := &BallPushEnv{
		Agent:    NewVerletParticle(pixel.ZV, 1, settings.DeltaTime),
		Ball:     NewVerletParticle(pixel.ZV, 1, settings.DeltaTime),
		Settings: settings,
		imd:      imdraw.New(nil),
	}
	e.Reset()
	return e
}

// ActionLength implements Env.
func (*BallPushEnv) ActionLength() int {
	return 2
}

// ConvertCategoricalAction implements Env.
func (*BallPushEnv) ConvertCategoricalAction(a int) []float64 {
	switch a {
	case 0:
		return []float64{1, 0}
	case 1:
		return []float64{-1, 0}
	case 2:
		return []float64{0, 1}
	case 3:
		return []float64{0, -1}
	case 4:
		return []float64{0, 0}
	}
	panic("invalid action")
}

// Name implements Env.
func (*BallPushEnv) Name() string {
	return "BallPush"
}

// NumCategoricalActions implements Env.
func (*BallPushEnv) NumCategoricalActions() int {
	return 5
}

// ObservationLength implements Env.
func (e *BallPushEnv) ObservationLength() int {
	return len(e.getObservation())
}

// Reset implements Env.
func (b *BallPushEnv) Reset() ResetData {
	b.Agent.SlideToPosition(pixel.V(0, rand.Float64()*b.Settings.BoundaryRadius*0.75).Rotated(rand.Float64() * 2 * math.Pi))
	b.Agent.SetVelocity(pixel.ZV)

	b.Ball.SlideToPosition(pixel.V(0, rand.Float64()*b.Settings.BoundaryRadius*0.75).Rotated(rand.Float64() * 2 * math.Pi))
	b.Ball.SetVelocity(pixel.ZV)

	b.HasTouchedBall = false
	b.HasCenteredBall = false

	return ResetData{
		Observation: b.getObservation(),
		Info:        map[string]interface{}{},
	}

}

func (b *BallPushEnv) getObservation() []float64 {
	// Things in the observation
	// 1. Vector from agent to center
	// 2. Vector from agent to ball
	// 3. Agent velocity
	// 4. Ball velocity
	return []float64{
		b.Agent.Position().X / b.Settings.BoundaryRadius,
		b.Agent.Position().Y / b.Settings.BoundaryRadius,
		b.Ball.Position().Sub(b.Agent.Position()).X / (2 * b.Settings.BoundaryRadius),
		b.Ball.Position().Sub(b.Agent.Position()).Y / (2 * b.Settings.BoundaryRadius),
		b.Agent.Velocity().X / b.Settings.AgentMaxSpeed(),
		b.Agent.Velocity().Y / b.Settings.AgentMaxSpeed(),
		b.Ball.Velocity().X / b.Settings.AgentMaxSpeed(),
		b.Ball.Velocity().Y / b.Settings.AgentMaxSpeed(),
	}
}

func (b *BallPushEnv) getInfo() map[string]interface{} {
	return make(map[string]interface{})
}

// Step implements Env.
func (e *BallPushEnv) Step(action []float64) StepData {
	validateAction(action, 2)

	justTouchedBall := false
	justCenteredBall := false

	// Ensure that going diagonally does not go faster than sideways
	controlVec := pixel.V(action[0], action[1])
	if controlVec.Len() > 0 {
		controlVec = controlVec.Unit()
	}

	agentControlForce := controlVec.Scaled(e.Settings.AgentAcceleration)
	agentDragForce := e.Agent.Velocity().Scaled(e.Settings.AgentDrag)
	e.Agent.ApplyForce(agentControlForce.Sub(agentDragForce))

	ballDragForce := e.Ball.Velocity().Scaled(e.Settings.BallDrag)
	e.Ball.ApplyForce(ballDragForce.Scaled(-1))

	boundaryOverlap := e.Agent.Position().Len() + e.Settings.AgentRadius - e.Settings.BoundaryRadius
	if boundaryOverlap > 0 {
		e.Agent.SlideToPosition(e.Agent.Position().Unit().Scaled(e.Settings.BoundaryRadius - e.Settings.AgentRadius))
	}

	ballOverlap := (e.Settings.AgentRadius + e.Settings.BallRadius) - e.Agent.Position().Sub(e.Ball.Position()).Len()
	if ballOverlap > 0 {
		if !e.HasTouchedBall {
			justTouchedBall = true
			e.HasTouchedBall = true
		}
		correctionVec := e.Agent.Position().Sub(e.Ball.Position()).Unit().Scaled(ballOverlap / 2)
		e.Agent.SlideToPosition(e.Agent.Position().Add(correctionVec))
		e.Ball.SlideToPosition(e.Ball.Position().Sub(correctionVec))
	}

	ballBoundaryOverlap := e.Ball.Position().Len() + e.Settings.BallRadius - e.Settings.BoundaryRadius
	if ballBoundaryOverlap > 0 {
		e.Ball.SlideToPosition(e.Ball.Position().Unit().Scaled(e.Settings.BoundaryRadius - e.Settings.BallRadius))
	}

	ballInCenter := e.BallInCenter()
	if ballInCenter && !e.HasCenteredBall {
		justCenteredBall = true
		e.HasCenteredBall = true
	}

	e.Agent.StepParticle()
	e.Ball.StepParticle()

	ballVelTowardsCenter := e.Ball.Velocity().Dot(e.Ball.Position().Unit().Scaled(-1))

	reward := 0.0
	if justTouchedBall {
		reward += e.Settings.TouchBallReward
	}
	if justCenteredBall {
		reward += e.Settings.PlaceInCenterReward
	}
	reward += e.Settings.MoveToCenterReward * ballVelTowardsCenter * e.Settings.DeltaTime / e.Settings.BoundaryRadius // when we divide by boundary radius and delta time, we get a reward of 1 by moving all the way from the boundary to the center

	if !e.HasTouchedBall {
		agentBallDir := e.Ball.Position().Sub(e.Agent.Position()).Unit()
		agentVelTowardsBall := e.Agent.Velocity().Dot(agentBallDir)
		reward += e.Settings.MoveToBallReward * agentVelTowardsBall * e.Settings.DeltaTime / (2 * e.Settings.BoundaryRadius)
	}

	return StepData{
		Observation: e.getObservation(),
		Reward:      reward,
		Terminated:  false,
		Info:        e.getInfo(),
	}
}

// RenderSize implements Env.
func (b *BallPushEnv) RenderSize() (float64, float64) {
	s := b.Settings.BoundaryRadius * 2 * b.Settings.Scale
	return s, s
}

// Render implements Env.
func (b *BallPushEnv) Render(target pixel.Target) {
	b.imd.Clear()
	// Clear the screen.
	b.imd.Color = pixel.RGB(0, 0, 0)
	b.imd.Push(pixel.V(0, 0), pixel.V(b.RenderSize()))
	b.imd.Rectangle(0)

	// Draw the bound circle.
	b.imd.Color = pixel.RGB(0.1, 0.1, 0.1)
	b.imd.Push(pixel.V(b.RenderSize()).Scaled(0.5))
	b.imd.Circle(b.Settings.BoundaryRadius*b.Settings.Scale, 0)
	b.imd.Color = pixel.RGB(0.4, 0.4, 0.4)
	b.imd.Push(pixel.V(b.RenderSize()).Scaled(0.5))
	b.imd.Circle(b.Settings.BoundaryRadius*b.Settings.Scale, 3)

	// Draw the target.
	if b.BallInCenter() {
		b.imd.Color = pixel.RGB(0.0, 0.9, 0.0)
	} else {
		b.imd.Color = pixel.RGB(0.0, 0.2, 0.8)
	}
	b.imd.Push(pixel.V(b.RenderSize()).Scaled(0.5))
	b.imd.Circle(b.Settings.TargetRadius*b.Settings.Scale, 3)

	// Draw the ball.
	b.imd.Color = pixel.RGB(0.0, 0.2, 0.8)
	b.imd.Push(b.Ball.Position().Scaled(b.Settings.Scale).Add(pixel.V(b.RenderSize()).Scaled(0.5)))
	b.imd.Circle(b.Settings.BallRadius*b.Settings.Scale, 0)

	// Draw the agent.
	b.imd.Color = pixel.RGB(0.8, 0.2, 0.0)
	b.imd.Push(b.Agent.Position().Scaled(b.Settings.Scale).Add(pixel.V(b.RenderSize()).Scaled(0.5)))
	b.imd.Circle(b.Settings.AgentRadius*b.Settings.Scale, 0)

	b.imd.Draw(target)
}
