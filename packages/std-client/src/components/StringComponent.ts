import { defineComponent, Metadata, Type, World } from "@mud-classic/recs";

export function defineStringComponent<M extends Metadata>(
  world: World,
  options?: { id?: string; metadata?: M; indexed?: boolean }
) {
  return defineComponent<{ value: Type.String }, M>(world, { value: Type.String }, options);
}
