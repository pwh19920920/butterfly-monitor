/**
 * @see https://umijs.org/docs/max/access#access
 */
export default function access(
  initialState:
    | {
        currentUser?: API.SysUser;
      }
    | undefined,
) {
  const { currentUser } = initialState ?? {};
  return {
    canAccess: (permission: string) =>
      !!currentUser && !!currentUser.permissions?.includes(permission),
    routeAccess: (route: { name?: string }) =>
      !!currentUser &&
      !!route?.name &&
      !!currentUser.codes?.includes(route.name),
  };
}
