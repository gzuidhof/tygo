package typelias

// Represent a duration that would be parsed with smth like `time.ParseDuration(...)`
type DurationString = string

type CacheConfig struct {
	Key string        `json:"key"`
	Ttl DurationString `json:"ttl"`
}
