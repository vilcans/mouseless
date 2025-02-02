package main

import (
	"github.com/bendahl/uinput"
	log "github.com/sirupsen/logrus"
)

type VirtualMouse struct {
	uinputMouse     uinput.Mouse
	isPressed       map[MouseButton]bool
	triggeredKeys   map[uint16]MouseButton
	moveFractionX   float64
	moveFractionY   float64
	scrollFractionX float64
	scrollFractionY float64
}

func NewVirtualMouse() (*VirtualMouse, error) {
	var err error
	v := VirtualMouse{
		isPressed:       make(map[MouseButton]bool),
		triggeredKeys:   make(map[uint16]MouseButton),
		scrollFractionX: 0.0,
		scrollFractionY: 0.0,
	}
	v.uinputMouse, err = uinput.CreateMouse("/dev/uinput", []byte("mouseless"))
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (v *VirtualMouse) ButtonPress(triggeredByKey uint16, button MouseButton) {
	var err error
	v.triggeredKeys[triggeredByKey] = button
	v.isPressed[button] = true
	log.Debug("Mouse: pressing %v", button)
	if button == ButtonLeft {
		err = v.uinputMouse.LeftPress()
	} else if button == ButtonMiddle {
		err = v.uinputMouse.MiddleClick()
	} else if button == ButtonRight {
		err = v.uinputMouse.RightPress()
	} else {
		log.Warnf("Mouse: unknown button: %v", button)
	}
	if err != nil {
		log.Warnf("Mouse: button press failed: %v", err)
	}
}

func (v *VirtualMouse) OriginalKeyUp(code uint16) {
	if button, ok := v.triggeredKeys[code]; ok {
		if pressed, ok := v.isPressed[button]; ok && pressed {
			var err error
			log.Debugf("Mouse: releasing %v", button)
			if button == ButtonLeft {
				err = v.uinputMouse.LeftRelease()
			} else if button == ButtonMiddle {
				// todo
			} else if button == ButtonRight {
				err = v.uinputMouse.RightRelease()
			} else {
				log.Warnf("Mouse: unknown button: %v", button)
			}
			if err != nil {
				log.Warnf("Mouse: button release failed: %v", err)
			}
			delete(v.isPressed, button)
		}
		delete(v.triggeredKeys, code)
	}
}

func (v *VirtualMouse) Scroll(x float64, y float64) {
	v.scrollFractionX += x
	v.scrollFractionY += y
	// move only the integer part
	var xInt = int32(v.scrollFractionX)
	var yInt = int32(v.scrollFractionY)
	v.scrollFractionX -= float64(xInt)
	v.scrollFractionY -= float64(yInt)
	if xInt != 0 {
		log.Debugf("Mouse: scroll horizontal: %v", xInt)
		err := v.uinputMouse.Wheel(true, xInt)
		if err != nil {
			log.Warnf("Mouse: scroll failed: %v", err)
		}
	}
	if yInt != 0 {
		log.Debugf("Mouse: scroll vertical: %v", yInt)
		err := v.uinputMouse.Wheel(false, -yInt)
		if err != nil {
			log.Warnf("Mouse: scroll failed: %v", err)
		}
	}
}

func (v *VirtualMouse) Move(x float64, y float64) {
	// this seems to be necessary so that the speed does not change on diagonal move
	if x != 0 && y != 0 {
		x *= 0.546
		y *= 0.546
	}
	v.moveFractionX += x
	v.moveFractionY += y
	// move only the integer part
	var xInt = int32(v.moveFractionX)
	var yInt = int32(v.moveFractionY)
	v.moveFractionX -= float64(xInt)
	v.moveFractionY -= float64(yInt)
	if xInt != 0 || yInt != 0 {
		log.Debugf("Mouse: move %v %v", xInt, yInt)
		err := v.uinputMouse.Move(xInt, yInt)
		if err != nil {
			log.Warnf("Mouse: move failed: %v", err)
		}
	}
}

func (v *VirtualMouse) Close() {
	v.uinputMouse.Close()
}
