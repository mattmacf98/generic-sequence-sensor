package genericsequencesensor

import (
	"context"
	"errors"
	"fmt"
	"sync"

	sensor "go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var (
	GenericSequenceSensor = resource.NewModel("mattmacf", "generic-sequence-sensor", "generic-sequence-sensor")
	errUnimplemented      = errors.New("unimplemented")

	validMethods = map[string]bool{
		"Readings":       true,
		"GetImages":      true,
		"JointPositions": true,
	}
)

func init() {
	resource.RegisterComponent(sensor.API, GenericSequenceSensor,
		resource.Registration[sensor.Sensor, *Config]{
			Constructor: newGenericSequenceSensorGenericSequenceSensor,
		},
	)
}

type ResourceConfig struct {
	ResourceName    string   `json:"resource_name"`
	Method          string   `json:"method"`
	SequenceCapHz   float64  `json:"sequence_cap_hz"`
	Tags            []string `json:"tags"`
}

type SequenceConfig struct {
	Resources []ResourceConfig `json:"resources"`
}

type Config struct {
	Sequences []SequenceConfig `json:"sequences"`
}

func (cfg *Config) Validate(path string) ([]string, []string, error) {
	for i, seq := range cfg.Sequences {
		for j, res := range seq.Resources {
			if res.ResourceName == "" {
				return nil, nil, fmt.Errorf("%s.sequences[%d].resources[%d]: resource_name must not be empty", path, i, j)
			}
			if !validMethods[res.Method] {
				return nil, nil, fmt.Errorf("%s.sequences[%d].resources[%d]: method %q must be one of Readings, GetImages, JointPositions", path, i, j, res.Method)
			}
		}
	}
	return nil, nil, nil
}

type genericSequenceSensorGenericSequenceSensor struct {
	resource.AlwaysRebuild
	resource.Named

	name   resource.Name
	logger logging.Logger
	cfg    *Config

	mu             sync.Mutex
	sequenceActive bool
	sequenceTag    string

	cancelCtx  context.Context
	cancelFunc func()
}

func newGenericSequenceSensorGenericSequenceSensor(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}
	return NewGenericSequenceSensor(ctx, deps, rawConf.ResourceName(), conf, logger)
}

func NewGenericSequenceSensor(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *Config, logger logging.Logger) (sensor.Sensor, error) {
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	s := &genericSequenceSensorGenericSequenceSensor{
		name:       name,
		logger:     logger,
		cfg:        conf,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}
	return s, nil
}

func (s *genericSequenceSensorGenericSequenceSensor) Name() resource.Name {
	return s.name
}

func (s *genericSequenceSensorGenericSequenceSensor) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	s.mu.Lock()
	active := s.sequenceActive
	tag := s.sequenceTag
	s.mu.Unlock()

	if !active {
		return map[string]interface{}{}, nil
	}

	var tags []interface{}
	if tag != "" {
		tags = []interface{}{tag}
	} else {
		tags = []interface{}{}
	}

	sequences := make([]interface{}, len(s.cfg.Sequences))
	for i, seq := range s.cfg.Sequences {
		resources := make([]interface{}, len(seq.Resources))
		for j, res := range seq.Resources {
			resources[j] = map[string]interface{}{
				"resource_name": res.ResourceName,
				"method":        res.Method,
			}
		}
		sequences[i] = map[string]interface{}{
			"sequence_tags": tags,
			"resources":     resources,
		}
	}

	var overrides []interface{}
	for _, seq := range s.cfg.Sequences {
		for _, res := range seq.Resources {
			resTags := make([]interface{}, len(res.Tags))
			for i, t := range res.Tags {
				resTags[i] = t
			}
			overrides = append(overrides, map[string]interface{}{
				"resource_name":        res.ResourceName,
				"method":               res.Method,
				"capture_frequency_hz": res.SequenceCapHz,
				"tags":                 resTags,
			})
		}
	}

	return map[string]interface{}{
		"sequences": sequences,
		"overrides": overrides,
	}, nil
}

func (s *genericSequenceSensorGenericSequenceSensor) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	command, ok := cmd["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command field must be a string")
	}

	switch command {
	case "start":
		tag, ok := cmd["sequence_tag"].(string)
		if !ok {
			return nil, fmt.Errorf("start command requires a sequence_tag string")
		}
		s.mu.Lock()
		s.sequenceActive = true
		s.sequenceTag = tag
		s.mu.Unlock()
		return map[string]interface{}{}, nil

	case "stop":
		s.mu.Lock()
		s.sequenceActive = false
		s.sequenceTag = ""
		s.mu.Unlock()
		return map[string]interface{}{}, nil

	default:
		return nil, fmt.Errorf("unknown command: %q", command)
	}
}

func (s *genericSequenceSensorGenericSequenceSensor) Status(ctx context.Context) (map[string]interface{}, error) {
	return nil, errUnimplemented
}

func (s *genericSequenceSensorGenericSequenceSensor) Close(context.Context) error {
	s.cancelFunc()
	return nil
}
