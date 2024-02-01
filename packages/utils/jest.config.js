/** @type {import('ts-jest').Options} */
export default {
  preset: "ts-jest",
  roots: ["tests"],
  testEnvironment: "node",
  transform: {
    '^.+\\.(js|jsx)$': 'babel-jest',
  },
};
