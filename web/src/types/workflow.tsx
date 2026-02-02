import { Node, Edge } from "reactflow";

export type DataSource = "Domain" | "GitHub";
export type Frequency = "2hr" | "4hr" | "6hr" | "12hr" | "1 day";

export type NodeType =
  | "trigger"
  | "gobuster"
  | "nikto"
  | "nmap"
  | "sqlmap"
  | "wpscan"
  | "owasp-vulnerabilities"
  | "secret-scan"
  | "dependency-check"
  | "semgrep-scan"
  | "container-scan"
  | "flow-chart"
  | "auto-fix"
  | "email"
  | "github-issue"
  | "slack"
  | "decision"
  | "estimate-cost"
  | "policy-check"
  | "generate-iac"
  | "drift-check"
  | "kube-bench"
  | "iac-scan"
  | "generate-docs";

export interface WorkflowNode extends Node {
  type: NodeType;
  data: any;
}

export type WorkflowEdge = Edge;

export interface TriggerData {
  dataSource: DataSource;
  url: string;
  frequency: Frequency;
}

export interface Workflow {
  id: string;
  name: string;
  createdAt: string;
  updatedAt: string;
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
}
