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
	ResourceName string `json:"resource_name"`
	Method       string `json:"method"`
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

	mu           sync.Mutex
	sequenceTags []string

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
	tags := make([]interface{}, len(s.sequenceTags))
	for i, t := range s.sequenceTags {
		tags[i] = t
	}
	s.mu.Unlock()

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

	return map[string]interface{}{
		"sequences": sequences,
	}, nil
}

func (s *genericSequenceSensorGenericSequenceSensor) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	if _, ok := cmd["get_sequence_tags"]; ok {
		s.mu.Lock()
		tags := make([]interface{}, len(s.sequenceTags))
		for i, t := range s.sequenceTags {
			tags[i] = t
		}
		s.mu.Unlock()
		return map[string]interface{}{"sequence_tags": tags}, nil
	}

	if val, ok := cmd["set_sequence_tags"]; ok {
		rawTags, ok := val.([]interface{})
		if !ok {
			return nil, fmt.Errorf("set_sequence_tags value must be a list of strings")
		}
		tags := make([]string, len(rawTags))
		for i, v := range rawTags {
			s, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("set_sequence_tags: element %d is not a string", i)
			}
			tags[i] = s
		}
		s.mu.Lock()
		s.sequenceTags = tags
		s.mu.Unlock()
		return map[string]interface{}{}, nil
	}

	return nil, fmt.Errorf("unknown command: %v", cmd)
}

func (s *genericSequenceSensorGenericSequenceSensor) Status(ctx context.Context) (map[string]interface{}, error) {
	return nil, errUnimplemented
}

func (s *genericSequenceSensorGenericSequenceSensor) Close(context.Context) error {
	s.cancelFunc()
	return nil
}
