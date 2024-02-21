package gym

import (
	"math"
	"math/rand"

	"github.com/gopxl/pixel"
	"github.com/gopxl/pixel/imdraw"
	"golang.org/x/image/colornames"

	b2 "github.com/ByteArena/box2d"
)

var _ Env = &WalkerEnv{}

type WalkerEnv struct {
	world    *b2.B2World
	player   *Player
	floor    *Box
	settings WalkerSettings
	rocks    []*Box
	imd      *imdraw.IMDraw
}

type WalkerSettings struct {
	PlayerLimbLength float64
	PlayerLimbWidth  float64
	PlayerBodyLength float64
	PlayerBodyHeight float64

	JointMaxAngle    float64
	JointMaxVelocity float64
	JointMaxTorque   float64

	StopOnFall bool
}

var DefaultWalkerSettings = WalkerSettings{
	PlayerLimbLength: 1,
	PlayerLimbWidth:  0.15,
	PlayerBodyLength: 2,
	PlayerBodyHeight: 0.25,
	JointMaxAngle:    math.Pi / 1.5,
	JointMaxVelocity: 5,
	JointMaxTorque:   15,
	StopOnFall:       false,
}

func NewWalkerEnv(settings WalkerSettings) *WalkerEnv {
	world := b2.MakeB2World(b2.B2Vec2{X: 0, Y: -9.81})
	player := NewPlayer(
		&world,
		settings.PlayerLimbLength, settings.PlayerLimbWidth, settings.PlayerBodyLength, settings.PlayerBodyHeight,
		settings.JointMaxTorque, settings.JointMaxAngle, settings.JointMaxVelocity,
	)
	floor := NewBox(&world, 100, 1, false, 1, 1, colornames.Black)
	floor.Body.SetTransform(b2.B2Vec2{X: 45}, 0)

	rocks := make([]*Box, 90)

	for i := range rocks {
		xr := rand.Float64()
		x := xr*90 + 10
		r := (rand.Float64()*0.8 + 0.2) * xr
		rocks[i] = NewBox(&world, r, r, false, 1, .3, colornames.Black)
		rocks[i].Body.SetTransform(b2.B2Vec2{X: x, Y: 0.5}, rand.Float64()*6)
	}

	return &WalkerEnv{
		world:    &world,
		player:   player,
		floor:    floor,
		imd:      imdraw.New(nil),
		settings: settings,
		rocks:    rocks,
	}
}

// ActionLength implements Env.
func (*WalkerEnv) ActionLength() int {
	// One for each joint
	return 4
}

// ObservationLength implements Env.
func (e *WalkerEnv) ObservationLength() int {
	return len(e.getObservation())
}

// ConvertCategoricalAction implements Env.
func (*WalkerEnv) ConvertCategoricalAction(int) []float64 {
	panic("unimplemented")
}

// NumCategoricalActions implements Env.
func (*WalkerEnv) NumCategoricalActions() int {
	panic("unimplemented")
}

// Name implements Env.
func (*WalkerEnv) Name() string {
	return "Walker Environment"
}

func (e *WalkerEnv) getObservation() []float64 {
	// Motor angles norm within ranges
	// Motor velocities within ranges
	// Body angle sin, body angle cos
	motorAngles := e.player.GetMotorAngles()
	motorVels := e.player.GetMotorVelocities()
	bodyAngle := e.player.Head.Body.GetAngle()
	return []float64{
		motorAngles[0] / e.player.MaxJointAngle,
		motorAngles[1] / e.player.MaxJointAngle,
		motorAngles[2] / e.player.MaxJointAngle,
		motorAngles[3] / e.player.MaxJointAngle,

		motorVels[0] / e.player.MaxJointVelocity,
		motorVels[1] / e.player.MaxJointVelocity,
		motorVels[2] / e.player.MaxJointVelocity,
		motorVels[3] / e.player.MaxJointVelocity,

		math.Sin(bodyAngle),
		math.Cos(bodyAngle),
	}
}

// Reset implements Env.
func (e *WalkerEnv) Reset() ResetData {
	e.player.Teleport(pixel.V(0, 4))
	return ResetData{
		Observation: e.getObservation(),
		Info:        make(map[string]interface{}),
	}
}

// Step implements Env.
func (e *WalkerEnv) Step(action []float64) StepData {
	e.player.SetMotorSpeeds(action[0], action[1], action[2], action[3])
	e.world.Step(1.0/60, 6, 2)

	headVx := e.player.Head.Body.GetLinearVelocity().X
	headPy := e.player.Head.Body.GetPosition().Y

	return StepData{
		Observation: e.getObservation(),
		Reward:      headVx * (1.0 / 60),
		Terminated:  headPy < 1.2 && e.settings.StopOnFall,
		Info:        make(map[string]interface{}),
	}
}

// Render implements Env.
func (e *WalkerEnv) Render(target pixel.Target) {
	e.imd.Clear()

	// Draw background
	e.imd.SetMatrix(pixel.IM)
	e.imd.Color = pixel.RGB(0.15, 0.15, 0.15)
	e.imd.Push(pixel.ZV, pixel.V(800, 800))
	e.imd.Rectangle(0)

	ppm := 25.0
	cwo := pixel.Vec(e.player.Head.Body.GetPosition()).Scaled(-1).Add(pixel.V(400, 200).Scaled(1.0 / ppm))

	// Draw start line
	e.imd.SetMatrix(pixel.IM.Moved(cwo).Scaled(pixel.ZV, ppm))
	e.imd.Color = colornames.White
	e.imd.Push(pixel.ZV, pixel.V(0, 100))
	e.imd.Line(0.2)

	// Draw floor
	for _, r := range e.rocks {
		r.Draw(e.imd, cwo, ppm)
	}
	e.floor.Draw(e.imd, cwo, ppm)
	e.player.Draw(e.imd, cwo, ppm)

	e.imd.Draw(target)
}

// RenderSize implements Env.
func (*WalkerEnv) RenderSize() (float64, float64) {
	return 800, 800
}

type Player struct {
	Head             *Box
	LThigh           *Box
	LShin            *Box
	RThigh           *Box
	RShin            *Box
	LHip             *b2.B2RevoluteJoint
	RHip             *b2.B2RevoluteJoint
	LKnee            *b2.B2RevoluteJoint
	RKnee            *b2.B2RevoluteJoint
	MaxJointVelocity float64
	MaxJointAngle    float64
	MaxJointTorque   float64
	LimbLength       float64
	BodyLength       float64
	BodyHeight       float64
}

func NewPlayer(world *b2.B2World, limbLength, limbWidth, bodyLength, bodyHeight, legTorque, jointMaxAngle, jointMaxVelocity float64) *Player {
	head := NewBox(world, bodyLength, bodyHeight, true, 1, 0.3, colornames.Orange)
	lThigh := NewBox(world, limbWidth, limbLength, true, 1, 0.3, colornames.Red)
	lShin := NewBox(world, limbWidth, limbLength, true, 1, 0.3, colornames.Red)
	rThigh := NewBox(world, limbWidth, limbLength, true, 1, 0.3, colornames.Blue)
	rShin := NewBox(world, limbWidth, limbLength, true, 1, 0.3, colornames.Blue)

	lHip := b2.MakeB2RevoluteJointDef()
	lHip.BodyA = head.Body
	lHip.BodyB = lThigh.Body
	lHip.CollideConnected = false
	lHip.LocalAnchorA = b2.B2Vec2{X: -bodyLength / 2, Y: -bodyHeight / 2}
	lHip.LocalAnchorB = b2.B2Vec2{Y: limbLength / 2}
	lHip.EnableLimit = true
	lHip.LowerAngle = -jointMaxAngle
	lHip.UpperAngle = jointMaxAngle
	lHip.EnableMotor = true
	lHip.MotorSpeed = 0
	lHip.MaxMotorTorque = legTorque
	lHipJ := world.CreateJoint(&lHip).(*b2.B2RevoluteJoint)

	rHip := b2.MakeB2RevoluteJointDef()
	rHip.BodyA = head.Body
	rHip.BodyB = rThigh.Body
	rHip.CollideConnected = false
	rHip.LocalAnchorA = b2.B2Vec2{X: bodyLength / 2, Y: -bodyHeight / 2}
	rHip.LocalAnchorB = b2.B2Vec2{Y: limbLength / 2}
	rHip.EnableLimit = true
	rHip.LowerAngle = -jointMaxAngle
	rHip.UpperAngle = jointMaxAngle
	rHip.EnableMotor = true
	rHip.MotorSpeed = 0
	rHip.MaxMotorTorque = legTorque
	rHipJ := world.CreateJoint(&rHip).(*b2.B2RevoluteJoint)

	lKnee := b2.MakeB2RevoluteJointDef()
	lKnee.BodyA = lThigh.Body
	lKnee.BodyB = lShin.Body
	lKnee.CollideConnected = false
	lKnee.LocalAnchorA = b2.B2Vec2{Y: -limbLength / 2}
	lKnee.LocalAnchorB = b2.B2Vec2{Y: limbLength / 2}
	lKnee.EnableLimit = true
	lKnee.LowerAngle = -jointMaxAngle
	lKnee.UpperAngle = jointMaxAngle
	lKnee.EnableMotor = true
	lKnee.MotorSpeed = 0
	lKnee.MaxMotorTorque = legTorque
	lKneeJ := world.CreateJoint(&lKnee).(*b2.B2RevoluteJoint)

	rKnee := b2.MakeB2RevoluteJointDef()
	rKnee.BodyA = rThigh.Body
	rKnee.BodyB = rShin.Body
	rKnee.CollideConnected = false
	rKnee.LocalAnchorA = b2.B2Vec2{Y: -limbLength / 2}
	rKnee.LocalAnchorB = b2.B2Vec2{Y: limbLength / 2}
	rKnee.EnableLimit = true
	rKnee.LowerAngle = -jointMaxAngle
	rKnee.UpperAngle = jointMaxAngle
	rKnee.EnableMotor = true
	rKnee.MotorSpeed = 0
	rKnee.MaxMotorTorque = legTorque
	rKneeJ := world.CreateJoint(&rKnee).(*b2.B2RevoluteJoint)

	p := &Player{
		head,
		lThigh, lShin, rThigh, rShin,
		lHipJ, rHipJ, lKneeJ, rKneeJ,
		jointMaxVelocity, jointMaxAngle, legTorque,
		limbLength, bodyLength, bodyHeight,
	}
	p.Teleport(pixel.ZV)
	return p
}

func (p *Player) Draw(imd *imdraw.IMDraw, cameraWorldOffset pixel.Vec, pixelsPerMeter float64) {
	p.Head.Draw(imd, cameraWorldOffset, pixelsPerMeter)
	p.LThigh.Draw(imd, cameraWorldOffset, pixelsPerMeter)
	p.LShin.Draw(imd, cameraWorldOffset, pixelsPerMeter)
	p.RThigh.Draw(imd, cameraWorldOffset, pixelsPerMeter)
	p.RShin.Draw(imd, cameraWorldOffset, pixelsPerMeter)

	imd.Color = colornames.White
	imd.SetMatrix(pixel.IM.Rotated(pixel.ZV, p.Head.Body.GetAngle()).Moved(pixel.Vec(p.Head.Body.GetPosition()).Add(cameraWorldOffset)).Scaled(pixel.ZV, pixelsPerMeter))
	imd.Push(pixel.V(-0.25, 0), pixel.V(0.25, 0))
	imd.Circle(0.25, 0)

	imd.Color = colornames.Black
	imd.SetMatrix(pixel.IM.Rotated(pixel.ZV, p.Head.Body.GetAngle()).Moved(pixel.Vec(p.Head.Body.GetPosition()).Add(cameraWorldOffset)).Scaled(pixel.ZV, pixelsPerMeter))
	imd.Push(pixel.V(-0.15, -0.05), pixel.V(0.15, -0.05))
	imd.Circle(0.1, 0)
}

func (p *Player) Teleport(pos pixel.Vec) {
	p.Head.Body.SetTransform(b2.B2Vec2(pos), 0)
	p.LThigh.Body.SetTransform(b2.B2Vec2(pos.Add(pixel.V(-p.BodyLength/2, -(p.BodyHeight/2+p.LimbLength/2)))), 0)
	p.RThigh.Body.SetTransform(b2.B2Vec2(pos.Add(pixel.V(p.BodyLength/2, -(p.BodyHeight/2+p.LimbLength/2)))), 0)
	p.LShin.Body.SetTransform(b2.B2Vec2(pos.Add(pixel.V(-p.BodyLength/2, -(p.BodyHeight/2+p.LimbLength*3/2)))), 0)
	p.RShin.Body.SetTransform(b2.B2Vec2(pos.Add(pixel.V(p.BodyLength/2, -(p.BodyHeight/2+p.LimbLength*3/2)))), 0)

	p.Head.Body.SetLinearVelocity(b2.B2Vec2{})
	p.Head.Body.SetAngularVelocity(0)
	p.LThigh.Body.SetLinearVelocity(b2.B2Vec2{})
	p.LThigh.Body.SetAngularVelocity(0)
	p.RThigh.Body.SetLinearVelocity(b2.B2Vec2{})
	p.RThigh.Body.SetAngularVelocity(0)
	p.LShin.Body.SetLinearVelocity(b2.B2Vec2{})
	p.LShin.Body.SetAngularVelocity(0)
	p.RShin.Body.SetLinearVelocity(b2.B2Vec2{})
	p.RShin.Body.SetAngularVelocity(0)

	p.Head.Body.SetAwake(true)
}

// Make sure the vals are between -1 and 1
func (p *Player) SetMotorSpeeds(lHip, rHip, lKnee, rKnee float64) {
	p.LHip.SetMotorSpeed(p.MaxJointVelocity * lHip)
	p.RHip.SetMotorSpeed(p.MaxJointVelocity * rHip)
	p.LKnee.SetMotorSpeed(p.MaxJointVelocity * lKnee)
	p.RKnee.SetMotorSpeed(p.MaxJointVelocity * rKnee)
}

func (p *Player) GetMotorAngles() []float64 {
	return []float64{
		p.LHip.GetJointAngle(),
		p.RHip.GetJointAngle(),
		p.LKnee.GetJointAngle(),
		p.RKnee.GetJointAngle(),
	}
}

func (p *Player) GetMotorVelocities() []float64 {
	return []float64{
		p.LHip.GetJointSpeed(),
		p.RHip.GetJointSpeed(),
		p.LKnee.GetJointSpeed(),
		p.RKnee.GetJointSpeed(),
	}
}
