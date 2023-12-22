package gym

import "github.com/gopxl/pixel"

// VerletParticle is a particle that uses the Verlet integration method for physics.
// It is designed to be used for physics in environments.
type VerletParticle struct {
	currentPosition    pixel.Vec
	previousPosition   pixel.Vec
	mass               float64
	currentForce       pixel.Vec
	currentImpulse     pixel.Vec
	recentVelocity     pixel.Vec
	recentAcceleration pixel.Vec
	dt                 float64
}

// NewVerletParticle creates a new VerletParticle with the given position, mass, and time step.
// The time step must be the same as the time step used in the environment.
func NewVerletParticle(position pixel.Vec, mass, dt float64) *VerletParticle {
	return &VerletParticle{
		currentPosition:    position,
		previousPosition:   position,
		mass:               mass,
		currentForce:       pixel.ZV,
		recentVelocity:     pixel.ZV,
		recentAcceleration: pixel.ZV,
		dt:                 dt,
	}
}

// Get the current position of the particle.
func (p *VerletParticle) Position() pixel.Vec {
	return p.currentPosition
}

// Get the velocity of the particle.
func (p *VerletParticle) Velocity() pixel.Vec {
	return p.recentVelocity
}

// Get the acceleration of the particle.
func (p *VerletParticle) Acceleration() pixel.Vec {
	return p.recentAcceleration
}

// Get the mass of the particle.
func (p *VerletParticle) Mass() float64 {
	return p.mass
}

// Will add a force to be applied on the next update step. This is a FORCE not an IMPULSE.
func (p *VerletParticle) ApplyForce(force pixel.Vec) {
	p.currentForce = p.currentForce.Add(force)
}

// Will add an impulse to be applied on the next update step. This is an IMPULSE not a FORCE.
func (p *VerletParticle) ApplyImpulse(impulse pixel.Vec) {
	p.currentImpulse = p.currentImpulse.Add(impulse)
}

// Will set the position of the particle to the given position. CAUTION: This will not update the previous position, meaning that this will also change the velocity.
func (p *VerletParticle) SlideToPosition(newPos pixel.Vec) {
	p.currentPosition = newPos
}

// Will set the velocity of the particle to the given velocity, by changing the previous position.
func (p *VerletParticle) SetVelocity(vel pixel.Vec) {
	p.previousPosition = p.currentPosition.Sub(vel.Scaled(p.dt))
}

// Step the particle forward in time by one time step.
func (p *VerletParticle) StepParticle() {
	totalForce := p.currentForce.Add(p.currentImpulse.Scaled(1 / p.dt))
	acceleration := totalForce.Scaled(1 / p.mass)
	p.recentAcceleration = acceleration
	nextPosition := p.currentPosition.Scaled(2).Sub(p.previousPosition).Add(acceleration.Scaled(p.dt * p.dt))
	p.recentVelocity = nextPosition.Sub(p.previousPosition).Scaled(0.5 / p.dt) // Sub one infront from one behind
	p.previousPosition = p.currentPosition
	p.currentPosition = nextPosition
	p.currentForce = pixel.ZV
	p.currentImpulse = pixel.ZV
}
