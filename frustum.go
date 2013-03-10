package u3d

import (
	"math"

	unum "github.com/metaleap/go-util/num"
)

type Frustum struct {
	Planes [6]unum.Vec3
	Axes   struct {
		X, Y, Z unum.Vec3
	}

	sphereFactor                              unum.Vec2
	aspectRatio, tanRadHalf, tanRadHalfAspect float64
}

func (me *Frustum) HasPoint(pos, point *unum.Vec3, zNear, zFar float64) bool {
	var axisPos float64
	pp := point.Sub(pos)
	if axisPos = pp.Dot(&me.Axes.Z); axisPos > zFar || axisPos < zNear {
		return false
	}
	halfHeight := axisPos * me.tanRadHalf
	if axisPos = pp.Dot(&me.Axes.Y); -halfHeight > axisPos || axisPos > halfHeight {
		return false
	}
	halfWidth := halfHeight * me.aspectRatio
	if axisPos = pp.Dot(&me.Axes.X); -halfWidth > axisPos || axisPos > halfWidth {
		return false
	}
	return true
}

func (me *Frustum) HasSphere(pos, center *unum.Vec3, radius, zNear, zFar float64) (fullyInside, intersect bool) {
	if radius == 0 {
		fullyInside, intersect = me.HasPoint(pos, center, zNear, zFar), false
		return
	}
	var axPos, z, d float64
	cp := center.Sub(pos)
	if axPos = cp.Dot(&me.Axes.Z); axPos > zFar+radius || axPos < zNear-radius {
		return
	}
	if axPos > zFar-radius || axPos < zNear+radius {
		intersect = true
	}

	z, d = axPos*me.tanRadHalfAspect, me.sphereFactor.X*radius
	if axPos = cp.Dot(&me.Axes.X); axPos > z+d || axPos < -z-d {
		intersect = false
		return
	}
	if axPos > z-d || axPos < -z+d {
		intersect = true
	}

	z, d = z/me.aspectRatio, me.sphereFactor.Y*radius
	if axPos = cp.Dot(&me.Axes.Y); axPos > z+d || axPos < -z-d {
		intersect = false
		return
	}
	if axPos > z-d || axPos < -z+d {
		intersect = true
	}
	fullyInside = !intersect
	return
}

func (me *Frustum) UpdateAxes(dir, upVector, upAxis *unum.Vec3) {
	me.Axes.Z = *dir
	me.Axes.Z.Negate()
	me.Axes.X.SetFrom(upVector)
	me.Axes.X.SetFromCross(&me.Axes.Z)
	me.Axes.X.Normalize()
	if upAxis == nil {
		me.Axes.Y.SetFromCrossOf(&me.Axes.Z, &me.Axes.X)
	} else {
		me.Axes.Y = *upAxis
	}
}

func (me *Frustum) UpdateRatio(fovYRadHalf, aspectRatio float64) {
	me.aspectRatio = aspectRatio
	me.tanRadHalf = math.Tan(fovYRadHalf)
	me.tanRadHalfAspect = me.tanRadHalf * aspectRatio
	me.sphereFactor.Y = 1 / math.Cos(fovYRadHalf)
	me.sphereFactor.X = 1 / math.Cos(math.Atan(me.tanRadHalfAspect))
}
