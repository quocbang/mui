/* eslint-disable */

export const protobufPackage = "sites";

export enum SubType {
  SUB_TYPE_UNSPECIFIED = 0,
  OPERATOR = 1,
  MATERIAL = 2,
  TOOL = 3,
  UNRECOGNIZED = -1,
}

export function subTypeFromJSON(object: any): SubType {
  switch (object) {
    case 0:
    case "SUB_TYPE_UNSPECIFIED":
      return SubType.SUB_TYPE_UNSPECIFIED;
    case 1:
    case "OPERATOR":
      return SubType.OPERATOR;
    case 2:
    case "MATERIAL":
      return SubType.MATERIAL;
    case 3:
    case "TOOL":
      return SubType.TOOL;
    case -1:
    case "UNRECOGNIZED":
    default:
      return SubType.UNRECOGNIZED;
  }
}

export function subTypeToJSON(object: SubType): string {
  switch (object) {
    case SubType.SUB_TYPE_UNSPECIFIED:
      return "SUB_TYPE_UNSPECIFIED";
    case SubType.OPERATOR:
      return "OPERATOR";
    case SubType.MATERIAL:
      return "MATERIAL";
    case SubType.TOOL:
      return "TOOL";
    case SubType.UNRECOGNIZED:
    default:
      return "UNRECOGNIZED";
  }
}

export enum Type {
  /** TYPE_UNSPECIFIED - TYPE_UNSPECIFIED is unspecified site. */
  TYPE_UNSPECIFIED = 0,
  /** CONTAINER - CONTAINER is a non-discrete resource site, e.g. oil tank. */
  CONTAINER = 1,
  /** SLOT - SLOT is a 1-1 resource site. */
  SLOT = 2,
  /** COLLECTION - COLLECTION is a resource collection site. */
  COLLECTION = 3,
  /** QUEUE - QUEUE is a one-at-a-time resource site. */
  QUEUE = 4,
  /** COLQUEUE - COLQUEUE is a collection pipeline. */
  COLQUEUE = 5,
  UNRECOGNIZED = -1,
}

export function typeFromJSON(object: any): Type {
  switch (object) {
    case 0:
    case "TYPE_UNSPECIFIED":
      return Type.TYPE_UNSPECIFIED;
    case 1:
    case "CONTAINER":
      return Type.CONTAINER;
    case 2:
    case "SLOT":
      return Type.SLOT;
    case 3:
    case "COLLECTION":
      return Type.COLLECTION;
    case 4:
    case "QUEUE":
      return Type.QUEUE;
    case 5:
    case "COLQUEUE":
      return Type.COLQUEUE;
    case -1:
    case "UNRECOGNIZED":
    default:
      return Type.UNRECOGNIZED;
  }
}

export function typeToJSON(object: Type): string {
  switch (object) {
    case Type.TYPE_UNSPECIFIED:
      return "TYPE_UNSPECIFIED";
    case Type.CONTAINER:
      return "CONTAINER";
    case Type.SLOT:
      return "SLOT";
    case Type.COLLECTION:
      return "COLLECTION";
    case Type.QUEUE:
      return "QUEUE";
    case Type.COLQUEUE:
      return "COLQUEUE";
    case Type.UNRECOGNIZED:
    default:
      return "UNRECOGNIZED";
  }
}

export enum ActionType {
  ADD = 0,
  REMOVE = 1,
  UNRECOGNIZED = -1,
}

export function actionTypeFromJSON(object: any): ActionType {
  switch (object) {
    case 0:
    case "ADD":
      return ActionType.ADD;
    case 1:
    case "REMOVE":
      return ActionType.REMOVE;
    case -1:
    case "UNRECOGNIZED":
    default:
      return ActionType.UNRECOGNIZED;
  }
}

export function actionTypeToJSON(object: ActionType): string {
  switch (object) {
    case ActionType.ADD:
      return "ADD";
    case ActionType.REMOVE:
      return "REMOVE";
    case ActionType.UNRECOGNIZED:
    default:
      return "UNRECOGNIZED";
  }
}
