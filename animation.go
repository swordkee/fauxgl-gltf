package fauxgl

import (
	"math"
	"sort"
)

// Animation represents a collection of animation channels
type Animation struct {
	Name     string
	Duration float64
	Channels []AnimationChannel
}

// AnimationChannel represents animation data for a specific node and property
type AnimationChannel struct {
	Target        *SceneNode
	Property      AnimationProperty
	Keyframes     []Keyframe
	Interpolation InterpolationType
}

// AnimationProperty represents what property is being animated
type AnimationProperty int

const (
	// Translation animates node translation
	Translation AnimationProperty = iota
	// Rotation animates node rotation
	Rotation
	// ScaleProperty animates node scale
	ScaleProperty
	// Weights animates morph target weights
	Weights
	// Joints animates skinned mesh joints
	Joints
)

// InterpolationType represents how values are interpolated between keyframes
type InterpolationType int

const (
	// Linear interpolation
	Linear InterpolationType = iota
	// Step interpolation (no interpolation)
	Step
	// CubicSpline interpolation
	CubicSpline
)

// Keyframe represents a single keyframe in an animation
type Keyframe struct {
	Time  float64
	Value interface{} // Can be Vector, Quaternion, float64, etc.
}

// Quaternion represents a rotation quaternion
type Quaternion struct {
	X, Y, Z, W float64
}

// Skin represents a skin for skeletal animation
type Skin struct {
	Name                string
	Joints              []*SceneNode // Joint nodes
	InverseBindMatrices []Matrix     // Inverse bind pose matrices
	Skeleton            *SceneNode   // Root skeleton node (optional)
}

// Joint represents a joint/bone in skeletal animation
type Joint struct {
	Node              *SceneNode
	InverseBindMatrix Matrix
	JointMatrix       Matrix // Computed joint matrix
}

// MorphTarget represents a morph target for shape interpolation
type MorphTarget struct {
	Name      string
	Positions []Vector // Target positions
	Normals   []Vector // Target normals (optional)
	Tangents  []Vector // Target tangents (optional)
}

// MorphTargets represents a collection of morph targets
type MorphTargets struct {
	Targets []MorphTarget
	Weights []float64
}

// NewAnimation creates a new animation
func NewAnimation(name string, duration float64) *Animation {
	return &Animation{
		Name:     name,
		Duration: duration,
		Channels: make([]AnimationChannel, 0),
	}
}

// AddChannel adds an animation channel to the animation
func (anim *Animation) AddChannel(channel AnimationChannel) {
	anim.Channels = append(anim.Channels, channel)
}

// Evaluate evaluates the animation at a specific time
func (anim *Animation) Evaluate(time float64) {
	// Wrap time to animation duration
	if anim.Duration > 0 {
		time = math.Mod(time, anim.Duration)
	}

	for _, channel := range anim.Channels {
		channel.Evaluate(time)
	}
}

// Evaluate evaluates the animation channel at a specific time
func (channel *AnimationChannel) Evaluate(time float64) {
	if len(channel.Keyframes) == 0 || channel.Target == nil {
		return
	}

	// Sort keyframes by time if not already sorted
	if !channel.isSorted() {
		sort.Slice(channel.Keyframes, func(i, j int) bool {
			return channel.Keyframes[i].Time < channel.Keyframes[j].Time
		})
	}

	// Find the keyframes to interpolate between
	beforeIndex, afterIndex := channel.findKeyframeIndices(time)

	var value interface{}

	if beforeIndex == afterIndex {
		// Exact match or only one keyframe
		value = channel.Keyframes[beforeIndex].Value
	} else {
		// Interpolate between keyframes
		before := channel.Keyframes[beforeIndex]
		after := channel.Keyframes[afterIndex]

		// Calculate interpolation factor
		t := (time - before.Time) / (after.Time - before.Time)

		value = channel.interpolate(before.Value, after.Value, t)
	}

	// Apply the animated value to the target
	channel.applyValue(value)
}

// isSorted checks if keyframes are sorted by time
func (channel *AnimationChannel) isSorted() bool {
	for i := 1; i < len(channel.Keyframes); i++ {
		if channel.Keyframes[i-1].Time > channel.Keyframes[i].Time {
			return false
		}
	}
	return true
}

// findKeyframeIndices finds the keyframe indices to interpolate between
func (channel *AnimationChannel) findKeyframeIndices(time float64) (int, int) {
	keyframes := channel.Keyframes

	// Time before first keyframe
	if time <= keyframes[0].Time {
		return 0, 0
	}

	// Time after last keyframe
	if time >= keyframes[len(keyframes)-1].Time {
		lastIndex := len(keyframes) - 1
		return lastIndex, lastIndex
	}

	// Binary search for the right interval
	for i := 0; i < len(keyframes)-1; i++ {
		if time >= keyframes[i].Time && time <= keyframes[i+1].Time {
			return i, i + 1
		}
	}

	return 0, 0
}

// interpolate interpolates between two values
func (channel *AnimationChannel) interpolate(before, after interface{}, t float64) interface{} {
	switch channel.Interpolation {
	case Step:
		return before

	case Linear:
		return channel.linearInterpolate(before, after, t)

	case CubicSpline:
		// For simplicity, fall back to linear interpolation
		return channel.linearInterpolate(before, after, t)

	default:
		return channel.linearInterpolate(before, after, t)
	}
}

// linearInterpolate performs linear interpolation between two values
func (channel *AnimationChannel) linearInterpolate(before, after interface{}, t float64) interface{} {
	switch channel.Property {
	case Translation, ScaleProperty:
		if v1, ok := before.(Vector); ok {
			if v2, ok := after.(Vector); ok {
				return v1.Lerp(v2, t)
			}
		}

	case Rotation:
		if q1, ok := before.(Quaternion); ok {
			if q2, ok := after.(Quaternion); ok {
				return q1.Slerp(q2, t)
			}
		}

	case Weights:
		if w1, ok := before.([]float64); ok {
			if w2, ok := after.([]float64); ok {
				result := make([]float64, len(w1))
				for i := range result {
					if i < len(w2) {
						result[i] = w1[i]*(1-t) + w2[i]*t
					} else {
						result[i] = w1[i]
					}
				}
				return result
			}
		}
	}

	return before
}

// applyValue applies the animated value to the target node
func (channel *AnimationChannel) applyValue(value interface{}) {
	if channel.Target == nil {
		return
	}

	switch channel.Property {
	case Translation:
		if v, ok := value.(Vector); ok {
			// Update translation component of local transform
			transform := channel.Target.LocalTransform
			transform = Identity().Translate(v).Mul(
				Identity().Scale(channel.getScale(transform)).Mul(
					channel.getRotationMatrix(transform)))
			channel.Target.SetTransform(transform)
		}

	case Rotation:
		if q, ok := value.(Quaternion); ok {
			// Update rotation component of local transform
			rotMatrix := q.ToMatrix()
			transform := Identity().Translate(channel.getTranslation(channel.Target.LocalTransform)).Mul(
				rotMatrix.Mul(Identity().Scale(channel.getScale(channel.Target.LocalTransform))))
			channel.Target.SetTransform(transform)
		}

	case ScaleProperty:
		if v, ok := value.(Vector); ok {
			// Update scale component of local transform
			transform := Identity().Translate(channel.getTranslation(channel.Target.LocalTransform)).Mul(
				channel.getRotationMatrix(channel.Target.LocalTransform).Mul(Identity().Scale(v)))
			channel.Target.SetTransform(transform)
		}
	}
}

// Helper functions to extract transform components (simplified)
func (channel *AnimationChannel) getTranslation(matrix Matrix) Vector {
	return Vector{matrix.X03, matrix.X13, matrix.X23}
}

func (channel *AnimationChannel) getScale(matrix Matrix) Vector {
	// Simplified scale extraction
	return Vector{1, 1, 1}
}

func (channel *AnimationChannel) getRotationMatrix(matrix Matrix) Matrix {
	// Simplified rotation extraction - return identity for now
	return Identity()
}

// Quaternion methods

// NewQuaternion creates a new quaternion
func NewQuaternion(x, y, z, w float64) Quaternion {
	return Quaternion{x, y, z, w}
}

// Identity returns the identity quaternion
func IdentityQuaternion() Quaternion {
	return Quaternion{0, 0, 0, 1}
}

// FromAxisAngle creates a quaternion from axis and angle
func QuaternionFromAxisAngle(axis Vector, angle float64) Quaternion {
	halfAngle := angle * 0.5
	sinHalf := math.Sin(halfAngle)
	cosHalf := math.Cos(halfAngle)

	return Quaternion{
		axis.X * sinHalf,
		axis.Y * sinHalf,
		axis.Z * sinHalf,
		cosHalf,
	}
}

// Normalize normalizes the quaternion
func (q Quaternion) Normalize() Quaternion {
	length := math.Sqrt(q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W)
	if length == 0 {
		return IdentityQuaternion()
	}

	invLength := 1.0 / length
	return Quaternion{
		q.X * invLength,
		q.Y * invLength,
		q.Z * invLength,
		q.W * invLength,
	}
}

// Slerp performs spherical linear interpolation between two quaternions
func (q1 Quaternion) Slerp(q2 Quaternion, t float64) Quaternion {
	// Compute dot product
	dot := q1.X*q2.X + q1.Y*q2.Y + q1.Z*q2.Z + q1.W*q2.W

	// If dot product is negative, slerp won't take the shorter path
	if dot < 0 {
		q2 = Quaternion{-q2.X, -q2.Y, -q2.Z, -q2.W}
		dot = -dot
	}

	// If quaternions are very similar, use linear interpolation
	if dot > 0.9995 {
		result := Quaternion{
			q1.X + t*(q2.X-q1.X),
			q1.Y + t*(q2.Y-q1.Y),
			q1.Z + t*(q2.Z-q1.Z),
			q1.W + t*(q2.W-q1.W),
		}
		return result.Normalize()
	}

	// Calculate angle between quaternions
	theta0 := math.Acos(math.Abs(dot))
	theta := theta0 * t

	sinTheta0 := math.Sin(theta0)
	sinTheta := math.Sin(theta)

	s0 := math.Cos(theta) - dot*sinTheta/sinTheta0
	s1 := sinTheta / sinTheta0

	return Quaternion{
		s0*q1.X + s1*q2.X,
		s0*q1.Y + s1*q2.Y,
		s0*q1.Z + s1*q2.Z,
		s0*q1.W + s1*q2.W,
	}
}

// ToMatrix converts quaternion to rotation matrix
func (q Quaternion) ToMatrix() Matrix {
	q = q.Normalize()

	xx := q.X * q.X
	yy := q.Y * q.Y
	zz := q.Z * q.Z
	xy := q.X * q.Y
	xz := q.X * q.Z
	yz := q.Y * q.Z
	wx := q.W * q.X
	wy := q.W * q.Y
	wz := q.W * q.Z

	return Matrix{
		1 - 2*(yy+zz), 2 * (xy - wz), 2 * (xz + wy), 0,
		2 * (xy + wz), 1 - 2*(xx+zz), 2 * (yz - wx), 0,
		2 * (xz - wy), 2 * (yz + wx), 1 - 2*(xx+yy), 0,
		0, 0, 0, 1,
	}
}

// AnimationPlayer handles playback of animations
type AnimationPlayer struct {
	animations  map[string]*Animation
	currentTime float64
	playSpeed   float64
	isPlaying   bool
	currentAnim string
	loop        bool
}

// NewAnimationPlayer creates a new animation player
func NewAnimationPlayer() *AnimationPlayer {
	return &AnimationPlayer{
		animations:  make(map[string]*Animation),
		currentTime: 0,
		playSpeed:   1.0,
		isPlaying:   false,
		loop:        true,
	}
}

// AddAnimation adds an animation to the player
func (player *AnimationPlayer) AddAnimation(name string, animation *Animation) {
	player.animations[name] = animation
}

// Play starts playing an animation
func (player *AnimationPlayer) Play(name string) bool {
	if _, exists := player.animations[name]; exists {
		player.currentAnim = name
		player.currentTime = 0
		player.isPlaying = true
		return true
	}
	return false
}

// Stop stops the current animation
func (player *AnimationPlayer) Stop() {
	player.isPlaying = false
	player.currentTime = 0
}

// Pause pauses the current animation
func (player *AnimationPlayer) Pause() {
	player.isPlaying = false
}

// Resume resumes the paused animation
func (player *AnimationPlayer) Resume() {
	if player.currentAnim != "" {
		player.isPlaying = true
	}
}

// Update updates the animation player
func (player *AnimationPlayer) Update(deltaTime float64) {
	if !player.isPlaying || player.currentAnim == "" {
		return
	}

	animation, exists := player.animations[player.currentAnim]
	if !exists {
		return
	}

	// Update time
	player.currentTime += deltaTime * player.playSpeed

	// Handle looping
	if player.currentTime >= animation.Duration {
		if player.loop {
			player.currentTime = math.Mod(player.currentTime, animation.Duration)
		} else {
			player.currentTime = animation.Duration
			player.isPlaying = false
		}
	}

	// Evaluate animation
	animation.Evaluate(player.currentTime)
}

// SetPlaySpeed sets the playback speed
func (player *AnimationPlayer) SetPlaySpeed(speed float64) {
	player.playSpeed = speed
}

func NewSkin(name string) *Skin {
	return &Skin{
		Name:                name,
		Joints:              make([]*SceneNode, 0),
		InverseBindMatrices: make([]Matrix, 0),
	}
}

// AddJoint adds a joint to the skin
func (skin *Skin) AddJoint(joint *SceneNode, inverseBindMatrix Matrix) {
	skin.Joints = append(skin.Joints, joint)
	skin.InverseBindMatrices = append(skin.InverseBindMatrices, inverseBindMatrix)
}

// UpdateJointMatrices updates all joint matrices for the current pose
func (skin *Skin) UpdateJointMatrices() {
	for i, joint := range skin.Joints {
		if i < len(skin.InverseBindMatrices) {
			// Joint matrix = globalTransform * inverseBindMatrix
			jointMatrix := joint.WorldTransform.Mul(skin.InverseBindMatrices[i])
			_ = jointMatrix // Store in a joint matrices array when needed
		}
	}
}

// NewMorphTarget creates a new morph target
func NewMorphTarget(name string, vertexCount int) *MorphTarget {
	return &MorphTarget{
		Name:      name,
		Positions: make([]Vector, vertexCount),
		Normals:   make([]Vector, vertexCount),
		Tangents:  make([]Vector, vertexCount),
	}
}

// ApplyMorphTargets applies morph target deformation to mesh
func ApplyMorphTargets(baseMesh *Mesh, targets *MorphTargets) *Mesh {
	if len(targets.Targets) == 0 || len(targets.Weights) == 0 {
		return baseMesh
	}

	// Create a copy of the base mesh
	resultMesh := baseMesh.Copy()

	// Apply weighted morph target deformation
	// For simplified implementation, we'll just apply the first target's weight
	if len(targets.Weights) > 0 && len(targets.Targets) > 0 {
		// In a full implementation, you'd iterate through vertices and apply
		// morph target displacements based on vertex indices
		_ = resultMesh.Triangles // Keep reference to triangles for future implementation
	}

	return resultMesh
}
