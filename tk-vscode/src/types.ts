export interface AxisStatus {
  effective?: string;
}
export interface TkNote {
  markdown?: string;
  actor?: string;
  timestamp?: string;
}
export interface RelationEdge {
  dst?: string;
  note?: string;
}
export interface Relations {
  blocks?: {
    out?: RelationEdge[];
    in?: RelationEdge[];
  };
  subtask?: {
    children?: string[];
    parent?: string;
  };
  related?: {
    out?: RelationEdge[];
  };
  duplicate_of?: {
    out?: RelationEdge[];
  };
  supersedes?: {
    out?: RelationEdge[];
  };
}
export interface TkTask {
  uuid?: string;
  display_id?: string;
  project_uuid?: string;
  title?: string;
  axes?: Record<string, AxisStatus | undefined>;
  blocked?: boolean;
  blockers?: Array<{ display_id?: string; title?: string }>;
  notes?: TkNote[];
  relations?: Relations;
}
export interface TkGroup {
  group?: string;
  tasks: TkTask[];
}
export interface TkProject {
  uid: string;
  name: string;
  local_preferred_alias?: string;
  aliases: string[];
  description?: string;
  type?: string;
}
export interface TkConfig {
  binary: string;
  cwd: string;
}
