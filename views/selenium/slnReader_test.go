package main

import (
	"testing"
	"runtime"
	"log"
)

func TestGetScenarioFilesList(t *testing.T) {

	_, err := GetScenarioFilesList()
	if err != nil {
		t.Error(err)
	}
}

func TestReadScenarioFiles(t *testing.T) {

	files, err := GetScenarioFilesList()
	if err != nil {
		t.Error(err)
	}
	for _, file := range files {
		_, err := ReadScenarioFiles(file)
		if err != nil {
			t.Error(err)
		}
	}
	runtime.NumGoroutine()
}

func TestGetDriverConnection(t *testing.T) {

	wd, err := GetDriverConnection()
	if err != nil {
		t.Error(err)
	}
	defer wd.Close()
}

func TestScenarioBufferHandler(t *testing.T) {

	scenarioBuffer := make(chan string)
	defer close(scenarioBuffer)
	files, err := GetScenarioFilesList()
	if err != nil {
		t.Error(err)
	}
	go func() {
		scenarioBuffer <- files[0]
	}()

	handlerResultsBuffer := make(chan HandlingResultType)
	defer close(handlerResultsBuffer)

	wd, err := GetDriverConnection()
	if err != nil {
		t.Error(err)
	}
	go ScenarioBufferHandler(scenarioBuffer, handlerResultsBuffer, wd)
	log.Println(<- handlerResultsBuffer)
	defer wd.Close()
}


