import type { INode } from "./node.model";
import type { IWorkspaceList } from "./workspace.model";
import type { INodeGpu } from "@/models/node-gpu.model";

export type TEntity = INode | IWorkspaceList | INodeGpu;
