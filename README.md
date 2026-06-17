# generic-sequence-sensor Module

The `mattmacf:generic-sequence-sensor` module provides a sensor that stores a set of named sequences — each a list of (resource, method) pairs — and manages capture-frequency overrides at runtime. Start a sequence to activate high-frequency capture on configured resources; stop it to zero out capture and clear the tag.

---

## Model: `mattmacf:generic-sequence-sensor:generic-sequence-sensor`

**API:** `rdk:component:sensor`

### Configuration

```json
{
  "sequences": [
    {
      "resources": [
        {
          "resource_name": "camera-1",
          "method": "GetImages",
          "sequence_cap_hz": 10,
          "tags": ["foo"]
        },
        {
          "resource_name": "arm-1",
          "method": "JointPositions",
          "sequence_cap_hz": 5,
          "tags": ["bar"]
        }
      ]
    }
  ]
}
```

| Name        | Type  | Required | Description                                                          |
| ----------- | ----- | -------- | -------------------------------------------------------------------- |
| `sequences` | array | Yes      | One or more sequence definitions. Each must have a `resources` list. |

Each entry in `resources`:

| Name                | Type     | Required | Description                                                                                   |
| ------------------- | -------- | -------- | --------------------------------------------------------------------------------------------- |
| `resource_name`     | string   | Yes      | Name of the resource involved in this step.                                                   |
| `method`            | string   | Yes      | Method to associate. Must be `Readings`, `GetImages`, or `JointPositions`.                    |
| `sequence_cap_hz`   | float    | No       | Capture frequency (Hz) to apply when the sequence is active. Defaults to `0`.                 |
| `tags`              | []string | No       | Data-capture tags to include in overrides for this resource.                                  |

### Readings

Returns all configured sequences annotated with the current active tag, plus a flat `overrides` list for every resource. `capture_frequency_hz` is the configured value while the sequence is running. When stopped (or before any sequence has been started), returns `{}`.

```json
{
  "sequences": [
    {
      "sequence_tags": ["my-tag"],
      "resources": [
        {"resource_name": "camera-1", "method": "GetImages"},
        {"resource_name": "arm-1",    "method": "JointPositions"}
      ]
    }
  ],
  "overrides": [
    {
      "resource_name": "camera-1",
      "method": "GetImages",
      "capture_frequency_hz": 10,
      "tags": ["foo"]
    },
    {
      "resource_name": "arm-1",
      "method": "JointPositions",
      "capture_frequency_hz": 5,
      "tags": ["bar"]
    }
  ]
}
```

### DoCommand

**`start`** — Activate the sequence and set the capture-frequency overrides to their configured values.

```json
{"command": "start", "sequence_tag": "my-tag"}
```

Returns `{}`.

---

**`stop`** — Deactivate the sequence. Clears the tag and causes `Readings` to return `{}`.

```json
{"command": "stop"}
```

Returns `{}`.
