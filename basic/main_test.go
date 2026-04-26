package main

import "testing"

func TestAddTruck(t *testing.T) {
	manager := NewTruckManager()
	err := manager.AddTruck("T-1", 100)

	if err != nil {
		t.Error(err)
	}
	if len(manager.trucks) != 1 {
		t.Errorf("AddTruck returned wrong number of trucks")
	}
}

func TestGetTruck(t *testing.T) {
	manager := NewTruckManager()
	_ = manager.AddTruck("T-1", 100)

	truck, err := manager.GetTruck("T-1")
	if err != nil {
		t.Error(err)
	}

	if truck.ID != "T-1" {
		t.Errorf("GetTruck returned wrong truck")
	}
	if truck.Cargo != 100 {
		t.Errorf("GetTruck returned wrong cargo")
	}
}

func TestRemoveTruck(t *testing.T) {
	manager := NewTruckManager()
	_ = manager.AddTruck("T-1", 100)
	_ = manager.AddTruck("T-2", 200)

	err := manager.RemoveTruck("T-1")
	if err != nil {
		t.Error(err)
	}
	if len(manager.trucks) != 1 {
		t.Errorf("RemoveTruck returned wrong number of trucks")
	}
}

func TestUpdateTruck(t *testing.T) {
	manager := NewTruckManager()
	_ = manager.AddTruck("T-1", 100)
	_ = manager.AddTruck("T-2", 200)

	err := manager.UpdateTruck("T-1", func(int) int { return 999 })
	if err != nil {
		t.Error(err)
	}
	truck, err := manager.GetTruck("T-1")
	if err != nil {
		t.Error(err)
	}
	if truck.ID != "T-1" {
		t.Errorf("GetTruck returned wrong truck")
	}
	if truck.Cargo != 999 {
		t.Errorf("GetTruck returned wrong cargo")
	}
}

func TestConcurrentUpdateTruck(t *testing.T) {
	manager := NewTruckManager()
	_ = manager.AddTruck("T-1", 100)

	const numGoroutines = 100
	const iterations = 100
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				_ = manager.UpdateTruck("T-1", func(c int) int { return c + 1 })
			}
			done <- true
		}()
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	expectedFinalValue := numGoroutines*iterations + 100
	finalTruck, _ := manager.GetTruck("T-1")

	if finalTruck.Cargo != expectedFinalValue {
		t.Errorf("Expected final cargo to be %d, got %d", expectedFinalValue, finalTruck.Cargo)
	}
}
