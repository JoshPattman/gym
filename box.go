package gym

import (
	"image/color"

	b2 "github.com/ByteArena/box2d"
	"github.com/gopxl/pixel"
	"github.com/gopxl/pixel/imdraw"
)

type Box struct {
	Body   *b2.B2Body
	width  float64
	height float64
	color  color.Color
}

func NewBox(world *b2.B2World, width, height float64, dynamic bool, density, friction float64, color color.Color) *Box {
	bodyDef := b2.MakeB2BodyDef()
	if dynamic {
		bodyDef.Type = b2.B2BodyType.B2_dynamicBody
	}
	bodyDef.Position = b2.B2Vec2{X: 0, Y: 0}
	body := world.CreateBody(&bodyDef)
	poly := b2.MakeB2PolygonShape()
	poly.SetAsBox(width/2, height/2)
	fixture := b2.MakeB2FixtureDef()
	fixture.Shape = &poly
	fixture.Density = density
	fixture.Friction = friction
	body.CreateFixtureFromDef(&fixture)
	return &Box{
		Body:   body,
		width:  width,
		height: height,
		color:  color,
	}
}

func (b *Box) Draw(imd *imdraw.IMDraw, cameraWorldOffset pixel.Vec, pixelsPerMeter float64) {
	imd.Color = b.color
	imd.SetMatrix(pixel.IM.Rotated(pixel.ZV, b.Body.GetAngle()).Moved(pixel.Vec(b.Body.GetPosition()).Add(cameraWorldOffset)).Scaled(pixel.ZV, pixelsPerMeter))

	imd.Push(pixel.V(-b.width/2, -b.height/2), pixel.V(b.width/2, b.height/2))

	imd.Rectangle(0)
}
