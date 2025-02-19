// Code generated by tygo. DO NOT EDIT.

//////////
// source: simple.go

/**
 * Comments are kept :)
 */
export type ComplexType = { [key: string]: { [key: number /* uint16 */]: number /* uint32 */ | undefined}};
export type UserRole = string;
export const UserRoleDefault: UserRole = "viewer";
export const UserRoleEditor: UserRole = "editor"; // Line comments are also kept
export interface UserEntry {
  /**
   * Instead of specifying `tstype` we could also declare the typing
   * for uuid.NullUUID in the config file.
   */
  id: string | null;
  prefs: { [key: string]: {
    foo: number /* uint32 */;
    /**
     * An unknown type without a `tstype` tag or mapping in the config file
     * uses the `fallback_type`, which defaults to `any`.
     */
    bar: unknown /* uuid.UUID */;
  }};
  address?: string;
  nickname?: string;
  role: UserRole;
  createdAt?: string /* RFC3339 */;
  complex: ComplexType;
}
export interface ListUsersResponse {
  users: UserEntry[];
}
