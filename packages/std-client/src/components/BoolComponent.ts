import { defineComponent, Metadata, Type, World } from "@mud-classic/recs";

export function defineBoolComponent<M extends Metadata>(
  world: World,
  options?: { id?: string; metadata?: M; indexed?: boolean }
) {
  return defineComponent<{ value: Type.Boolean }, M>(world, { value: Type.Boolean }, options);
}
