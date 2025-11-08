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
  title?: string;
  axes?: Record<string, AxisStatus | undefined>;
  blocked?: boolean;
  blockers?: Array<{ display_id?: string; title?: string }>;
  notes?: TkNote[];
  project_uuid?: string;
  created_at?: string;
  created_by?: string;
  metadata?: Record<string, any>;
  relations?: Relations;
}

export interface VSCodeAPI {
  postMessage(message: any): void;
  getState(): any;
  setState(state: any): void;
}
