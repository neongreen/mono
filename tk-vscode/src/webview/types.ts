export interface AxisStatus {
  effective?: string;
}

export interface TkNote {
  markdown?: string;
  actor?: string;
  timestamp?: string;
}

export interface TkTask {
  task_uuid?: string;
  task_id?: string;
  title?: string;
  axes?: Record<string, AxisStatus | undefined>;
  blocked?: boolean;
  blockers?: Array<{ task_id?: string; title?: string }>;
  notes?: TkNote[];
}

export interface VSCodeAPI {
  postMessage(message: any): void;
  getState(): any;
  setState(state: any): void;
}
