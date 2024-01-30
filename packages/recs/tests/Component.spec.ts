import {
  defineComponent,
  setComponent,
  removeComponent,
  getComponentValue,
  hasComponent,
  withValue,
  componentValueEquals,
  getEntitiesWithValue,
} from "../src/Component";
import { Type } from "../src/constants";
import { createEntity } from "../src/Entity";
import { AnyComponent, EntityID, EntityIndex, World } from "../src/types";
import { createWorld } from "../src/World";

describe("Component", () => {
  let world: World;

  beforeEach(() => {
    world = createWorld();
  });

  it("emit changes to its stream", () => {
    const entity = createEntity(world);
    const component = defineComponent(world, { x: Type.Number, y: Type.Number });

    const mock = jest.fn();
    component.update$.subscribe((update) => {
      mock(update);
    });

    setComponent(component, entity, { x: 1, y: 2 });
    setComponent(component, entity, { x: 7, y: 2 });
    setComponent(component, entity, { x: 7, y: 2 });
    removeComponent(component, entity);

    expect(mock).toHaveBeenNthCalledWith(1, { entity, value: [{ x: 1, y: 2 }, undefined], component });
    expect(mock).toHaveBeenNthCalledWith(2, {
      entity,
      component,
      value: [
        { x: 7, y: 2 },
        { x: 1, y: 2 },
      ],
    });
    expect(mock).toHaveBeenNthCalledWith(3, {
      entity,
      component,
      value: [
        { x: 7, y: 2 },
        { x: 7, y: 2 },
      ],
    });
    expect(mock).toHaveBeenNthCalledWith(4, { entity, component, value: [undefined, { x: 7, y: 2 }] });
  });

  describe("defineComponent", () => {
    it("should register the component in the world", () => {
      expect(world.components.length).toBe(0);
      defineComponent(world, { value: Type.Boolean });
      expect(world.components.length).toBe(1);
    });
  });

  describe("setComponent", () => {
    let component: AnyComponent;
    let entity: EntityIndex;
    let value: number;

    beforeEach(() => {
      component = defineComponent(world, { value: Type.Number });
      entity = createEntity(world);
      value = 1;
      setComponent(component, entity, { value });
    });

    it("should store the component value", () => {
      expect(component.values.value.get(entity)).toBe(value);
    });

    it("should store the entity", () => {
      expect(hasComponent(component, entity)).toBe(true);
    });

    it.todo("should store the value array");
  });

  describe("removeComponent", () => {
    let component: AnyComponent;
    let entity: EntityIndex;
    let value: number;

    beforeEach(() => {
      component = defineComponent(world, { value: Type.Number });
      entity = createEntity(world);
      value = 1;
      setComponent(component, entity, { value });
      removeComponent(component, entity);
    });

    it("should remove the component value", () => {
      expect(component.values.value.get(entity)).toBe(undefined);
    });

    it("should remove the entity", () => {
      expect(hasComponent(component, entity)).toBe(false);
    });

    // it("shouldremove the component from the entity's component set", () => {
    //   expect(world.entities.get(entity)?.has(component)).toBe(false);
    // });
  });

  describe("hasComponent", () => {
    it("should return true if the entity has the component", () => {
      const component = defineComponent(world, { x: Type.Number, y: Type.Number });
      const entity = createEntity(world);
      const value = { x: 1, y: 2 };
      setComponent(component, entity, value);

      expect(hasComponent(component, entity)).toEqual(true);
    });
  });

  describe("getComponentValue", () => {
    it("should return the correct component value", () => {
      const component = defineComponent(world, { x: Type.Number, y: Type.Number });
      const entity = createEntity(world);
      const value = { x: 1, y: 2 };
      setComponent(component, entity, value);

      const receivedValue = getComponentValue(component, entity);
      expect(receivedValue).toEqual(value);
    });
  });

  describe("getComponentValueStrict", () => {
    it.todo("should return the correct component value");
    it.todo("should error if the component value does not exist");
  });

  describe("componentValueEquals", () => {
    const value1 = { x: 1, y: 2, z: "x" };
    const value2 = { x: 1, y: 2, z: "x" };
    const value3 = { x: "1", y: 2, z: "x" };

    expect(componentValueEquals(value1, value2)).toBe(true);
    expect(componentValueEquals(value2, value3)).toBe(false);
  });

  describe("withValue", () => {
    it("should return a ComponentWithValue", () => {
      const component = defineComponent(world, { x: Type.Number, y: Type.Number });
      const value = { x: 1, y: 2 };
      const componentWithValue = withValue(component, value);
      expect(componentWithValue).toEqual([component, value]);
    });
  });

  describe("getEntitiesWithValue", () => {
    it("Should return all and only entities with this value", () => {
      const Position = defineComponent(world, { x: Type.Number, y: Type.Number });
      const entity1 = createEntity(world, [withValue(Position, { x: 1, y: 2 })]);
      createEntity(world, [withValue(Position, { x: 2, y: 1 })]);
      createEntity(world);
      const entity4 = createEntity(world, [withValue(Position, { x: 1, y: 2 })]);

      expect(getEntitiesWithValue(Position, { x: 1, y: 2 })).toEqual(new Set([entity1, entity4]));
    });
  });
});
