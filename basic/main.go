package main

import (
	"errors"
	"sync"
)

var ErrTruckNotFound = errors.New("truck not found")

type FleetManager interface {
	AddTruck(id string, cargon int) error
	GetTruck(id string) (Truck, error)
	RemoveTruck(id string) error
	UpdateTruck(id string, cargo func(int) int) error
}

type Truck struct {
	ID    string
	Cargo int
}

type TruckManager struct {
	trucks map[string]*Truck
	sync.RWMutex
}

func NewTruckManager() TruckManager {
	return TruckManager{
		trucks: make(map[string]*Truck),
	}
}

func (tm *TruckManager) AddTruck(id string, cargon int) error {
	tm.Lock()
	defer tm.Unlock()

	tm.trucks[id] = &Truck{
		ID:    id,
		Cargo: cargon,
	}
	return nil
}

func (tm *TruckManager) GetTruck(id string) (Truck, error) {
	tm.RLock()
	defer tm.RUnlock()

	truck, ok := tm.trucks[id]
	if !ok {
		return Truck{}, ErrTruckNotFound
	}
	return *truck, nil
}

func (tm *TruckManager) RemoveTruck(id string) error {
	tm.Lock()
	defer tm.Unlock()

	_, ok := tm.trucks[id]
	if !ok {
		return ErrTruckNotFound
	}
	delete(tm.trucks, id)
	return nil
}

func (tm *TruckManager) UpdateTruck(id string, cargo func(int) int) error {
	tm.Lock()
	defer tm.Unlock()

	truck, ok := tm.trucks[id]
	if !ok {
		return ErrTruckNotFound
	}

	truck.Cargo = cargo(truck.Cargo)
	return nil
}
