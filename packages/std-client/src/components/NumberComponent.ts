import { defineComponent, Metadata, Type, World } from "@mud-classic/recs";

export function defineNumberComponent<M extends Metadata>(
  world: World,
  options?: { id?: string; metadata?: M; indexed?: boolean }
) {
  return defineComponent<{ value: Type.Number }, M>(world, { value: Type.Number }, options);
}
