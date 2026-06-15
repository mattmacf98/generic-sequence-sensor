# generic-sequence-sensor Module

The `mattmacf:generic-sequence-sensor` module provides a sensor that stores a set of named sequences — each a list of (resource, method) pairs — and lets you tag them at runtime via DoCommand. Readings return the full sequence list annotated with the current tags, making it easy to record what data was captured and under what conditions.

---

## Model: `mattmacf:generic-sequence-sensor:generic-sequence-sensor`

**API:** `rdk:component:sensor`

Stores a fixed set of sequence definitions (configured as JSON attributes) and a mutable list of sequence tags managed at runtime through `DoCommand`. `Readings` returns all sequences annotated with the current tags.

### Configuration

```json
{
  "sequences": [
    {
      "resources": [
        {"resource_name": "camera-1", "method": "GetImages"},
        {"resource_name": "arm-1",    "method": "JointPositions"}
      ]
    }
  ]
}
```

| Name        | Type  | Required | Description                                                          |
| ----------- | ----- | -------- | -------------------------------------------------------------------- |
| `sequences` | array | Yes      | One or more sequence definitions. Each must have a `resources` list. |

Each entry in `resources`:

| Name            | Type   | Required | Description                                                                          |
| --------------- | ------ | -------- | ------------------------------------------------------------------------------------ |
| `resource_name` | string | Yes      | Name of the resource (component or service) involved in this step.                   |
| `method`        | string | Yes      | Method to associate with this resource. Must be `Readings`, `GetImages`, or `JointPositions`. |

Sequence tags are **not** part of the config — they are set at runtime via `DoCommand` and default to an empty list on startup.

### Readings

Returns all configured sequences, each annotated with the current in-memory sequence tags:

```json
{
  "sequences": [
    {
      "sequence_tags": ["walking-demo"],
      "resources": [
        {"resource_name": "camera-1", "method": "GetImages"},
        {"resource_name": "arm-1",    "method": "JointPositions"}
      ]
    }
  ]
}
```

`sequence_tags` is empty (`[]`) until set via `DoCommand`.

### DoCommand

**`get_sequence_tags`** — Return the current sequence tags.

```json
{ "get_sequence_tags": true }
```

Returns:

```json
{ "sequence_tags": ["walking-demo"] }
```

---

**`set_sequence_tags`** — Replace the sequence tags list.

```json
{ "set_sequence_tags": ["walking-demo", "trial-3"] }
```

Returns:

```json
{}
```

Pass an empty list to clear all tags:

```json
{ "set_sequence_tags": [] }
```
