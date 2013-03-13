package u3d

import (
	"math"

	unum "github.com/metaleap/go-util/num"
)

type FrustumCoords struct {
	unum.Vec2
	C, TL, TR, BL, BR unum.Vec3

	x, y unum.Vec3
}

type Frustum struct {
	Bounding Bounds
	Planes   [6]FrustumPlane
	Axes     struct {
		X, Y, Z unum.Vec3
	}
	Near, Far FrustumCoords

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

func (me *Frustum) UpdateCoords(persp *Perspective, pos, dir, upVector, upAxis *unum.Vec3) {
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

	me.Near.C.SetFromSubScaled(pos, &me.Axes.Z, persp.ZNear)
	me.Far.C.SetFromSubScaled(pos, &me.Axes.Z, persp.ZFar)
	me.Near.y.SetFromScaled(&me.Axes.Y, me.Near.Y)
	me.Far.y.SetFromScaled(&me.Axes.Y, me.Far.Y)
	me.Near.x.SetFromScaled(&me.Axes.X, me.Near.X)
	me.Far.x.SetFromScaled(&me.Axes.X, me.Far.X)

	// ntl = nc + ny - nx
	me.Near.TL.SetFromAddSub(&me.Near.C, &me.Near.y, &me.Near.x)
	// ntr = nc + ny + nx
	me.Near.TR.SetFromAddAdd(&me.Near.C, &me.Near.y, &me.Near.x)
	// nbl = nc - ny - nx
	me.Near.BL.SetFromSubSub(&me.Near.C, &me.Near.y, &me.Near.x)
	// nbr = nc - ny + nx
	me.Near.BR.SetFromSubAdd(&me.Near.C, &me.Near.y, &me.Near.x)
	// ftl = fc + fy - fx
	me.Far.TL.SetFromAddSub(&me.Far.C, &me.Far.y, &me.Far.x)
	// fbr = fc - fy + fx
	me.Far.BR.SetFromSubAdd(&me.Far.C, &me.Far.y, &me.Far.x)
	// ftr = fc + fy + fx
	me.Far.TR.SetFromAddAdd(&me.Far.C, &me.Far.y, &me.Far.x)
	// fbl = fc - fy - fx
	me.Far.BL.SetFromSubSub(&me.Far.C, &me.Far.y, &me.Far.x)

	//	left
	me.Planes[0].setFrom(&me.Near.TL, &me.Near.BL, &me.Far.BL)
	//	right
	me.Planes[1].setFrom(&me.Near.BR, &me.Near.TR, &me.Far.BR)
	//	bottom
	me.Planes[2].setFrom(&me.Near.BL, &me.Near.BR, &me.Far.BR)
	//	top
	me.Planes[3].setFrom(&me.Near.TR, &me.Near.TL, &me.Far.TL)
	//	near
	me.Planes[4].setFrom(&me.Near.TL, &me.Near.TR, &me.Near.BR)
	//	far
	me.Planes[5].setFrom(&me.Far.TR, &me.Far.TL, &me.Far.BL)
}

func (me *Frustum) UpdateRatio(persp *Perspective, aspectRatio float64) {
	me.aspectRatio = aspectRatio
	me.tanRadHalf = math.Tan(persp.FovY.RadHalf)
	me.tanRadHalfAspect = me.tanRadHalf * aspectRatio
	me.sphereFactor.Y = 1 / math.Cos(persp.FovY.RadHalf)
	me.sphereFactor.X = 1 / math.Cos(math.Atan(me.tanRadHalfAspect))
	me.Near.Y = persp.ZNear * me.tanRadHalf
	me.Near.X = me.Near.Y * aspectRatio
	me.Far.Y = persp.ZFar * me.tanRadHalf
	me.Far.X = me.Far.Y * aspectRatio
}

type FrustumPlane struct {
	Normal unum.Vec3
	D      float64
}

func (me *FrustumPlane) setFrom(p1, p2, p3 *unum.Vec3) {
	me.Normal.SetFromCrossOf(p3.Sub(p2), p1.Sub(p2))
	me.Normal.Normalize()
	me.D = -me.Normal.Dot(p2)
}
