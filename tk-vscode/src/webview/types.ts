export interface AxisStatus {
  effective?: string;
}

export interface TkNote {
  markdown?: string;
  actor?: string;
  timestamp?: string;
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
}

export interface VSCodeAPI {
  postMessage(message: any): void;
  getState(): any;
  setState(state: any): void;
}
