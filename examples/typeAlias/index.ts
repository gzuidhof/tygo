// Code generated by tygo. DO NOT EDIT.
export type TsDurationString = `${number}ms` | `${number}s` | `${number}m` | `${number}h`;
//////////
// source: alias.go

/**
 * Represent a duration that would be parsed with smth like `time.ParseDuration(...)`
 */
export type DurationString = string;
export interface CacheConfig {
  key: string;
  ttl: TsDurationString;
}
