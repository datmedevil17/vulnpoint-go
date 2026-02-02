import { useState, useEffect } from "react";
import { X, Minimize2, Maximize2, CheckCircle2, Circle, Clock, ChevronRight, ChevronDown } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";
import { workflowApi } from "@/hooks/useWorkflow";
import { Node } from "reactflow";

interface Step {
  id: string;
  name: string;
  status: "pending" | "running" | "completed" | "failed" | "skipped";
  duration?: string;
  output?: string;
  logs: string[];
}

interface WorkflowTerminalProps {
  open: boolean;
  onClose: () => void;
  workflowId?: string;
  executionId: string | null;
  nodes: Node[];
}

export function WorkflowTerminal({ open, onClose, executionId, nodes }: WorkflowTerminalProps) {
  const [minimized, setMinimized] = useState(false);
  const [expandedStep, setExpandedStep] = useState<string | null>(null);
  const [steps, setSteps] = useState<Step[]>([]);
  const [executionStatus, setExecutionStatus] = useState<string>("pending");
  const [progress, setProgress] = useState(0);

  // Poll execution status
  useEffect(() => {
    if (!open || !executionId) return;

    // Initial setup from nodes
    if (steps.length === 0 && nodes.length > 0) {
       const initialSteps = nodes.map(node => ({
           id: node.id,
           name: node.data?.label || node.type || node.id,
           status: "pending" as const,
           logs: []
       }));
       // Try to sort steps if possible, otherwise use node order
       // Ideally we would topo sort here, but for now we list them
       setSteps(initialSteps);
    }

    const poll = async () => {
      try {
        const execution = await workflowApi.getExecution(executionId);
        setExecutionStatus(execution.data.status);

        if (execution.data.results) {
            setSteps(prevSteps => {
                let completedCount = 0;
                const newSteps = prevSteps.map(step => {
                    const result = execution.data.results[step.id];
                    let status: Step["status"] = "pending";
                    let logs: string[] = [];
                    let output: string | undefined;

                    if (result) {
                        status = result.status === 'completed' ? 'completed' : 
                                 result.status === 'skipped' ? 'skipped' : 'failed';
                        if (result.output) {
                            if (typeof result.output === 'string') {
                                logs = result.output.split('\n');
                            } else {
                                output = JSON.stringify(result.output, null, 2);
                            }
                        }
                        // Also check for data object
                        if (result.data) {
                             output = JSON.stringify(result.data, null, 2);
                        }
                        
                        if (status === 'completed' || status === 'skipped') completedCount++;
                    } else if (execution.data.currentNode === step.id) {
                        status = 'running';
                    }

                    return {
                        ...step,
                        status,
                        logs,
                        output
                    };
                });
                
                setProgress(Math.round((completedCount / newSteps.length) * 100));
                return newSteps;
            });
        }

        if (execution.data.status === 'completed' || execution.data.status === 'failed') {
             // Stop polling? No, user might want to see it. But we can slow down.
        }
      } catch (error) {
        console.error("Failed to poll execution:", error);
      }
    };

    const interval = setInterval(poll, 1000); // Poll every 1s
    poll(); // Initial call

    return () => clearInterval(interval);
  }, [open, executionId, nodes]); // Dependencies

  if (!open) return null;

  return (
    <div className={cn(
      "fixed bottom-4 right-4 z-50 transition-all duration-300 ease-in-out shadow-2xl border border-border/50 rounded-xl bg-zinc-950/95 text-zinc-50 backdrop-blur-xl overflow-hidden font-mono",
      minimized ? "w-72 h-14" : "w-[600px] h-[500px]"
    )}>
      {/* Header */}
      <div className="flex items-center justify-between p-3 bg-zinc-900/50 border-b border-white/5 cursor-pointer" onClick={() => setMinimized(!minimized)}>
            <div className="flex items-center gap-2">
                <div className={`w-2 h-2 rounded-full ${executionStatus === 'running' ? "animate-pulse bg-green-500" : executionStatus === 'completed' ? "bg-blue-500" : "bg-zinc-700"}`} />
                <span className="text-xs font-semibold tracking-wide text-zinc-400">
                    {minimized ? (executionStatus === 'running' ? "Running Pipeline..." : "Pipeline Finished") : "EXECUTION PROGRESS"}
                </span>
                {executionId && <span className="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-500">ID: {executionId.slice(0, 8)}</span>}
            </div>
            <div className="flex items-center gap-1">
                <button 
                    onClick={(e) => { e.stopPropagation(); setMinimized(!minimized); }}
                    className="p-1 hover:bg-white/10 rounded"
                >
                    {minimized ? <Maximize2 className="w-3 h-3" /> : <Minimize2 className="w-3 h-3" />}
                </button>
                <button 
                    onClick={(e) => { e.stopPropagation(); onClose(); }}
                    className="p-1 hover:bg-red-500/20 text-red-400 rounded"
                >
                    <X className="w-3 h-3" />
                </button>
            </div>
      </div>

      {!minimized && (
        <div className="flex flex-col h-[calc(100%-48px)]">
            {/* Progress Bar */}
            <div className="px-6 py-4 border-b border-white/5 bg-zinc-900/20">
                <div className="flex justify-between text-xs mb-2">
                    <span className={cn("font-semibold px-2 py-0.5 rounded", 
                        executionStatus === 'running' ? "text-blue-400 bg-blue-500/10" : 
                        executionStatus === 'completed' ? "text-green-400 bg-green-500/10" : "text-zinc-500"
                    )}>
                        {executionStatus.toUpperCase()}
                    </span>
                    <span className="text-zinc-500">{steps.filter(s => s.status === 'completed' || s.status === 'skipped').length} of {steps.length} steps completed</span>
                </div>
                
                {/* Stepper */}
                <div className="flex items-center justify-between mt-4 relative">
                    <div className="absolute left-0 top-1/2 -translate-y-1/2 w-full h-0.5 bg-zinc-800 -z-10" />
                    <div className="absolute left-0 top-1/2 -translate-y-1/2 h-0.5 bg-green-500 -z-10 transition-all duration-500" style={{ width: `${progress}%` }} />
                    
                    {steps.map((step, i) => (
                        <div key={step.id} className="flex flex-col items-center gap-2 group">
                            <div className={cn(
                                "w-6 h-6 rounded-full flex items-center justify-center border-2 text-[10px] font-bold transition-all duration-300 bg-zinc-950",
                                step.status === 'completed' ? "border-green-500 text-green-500" :
                                step.status === 'running' ? "border-blue-500 text-blue-500 shadow-[0_0_10px_rgba(59,130,246,0.5)]" :
                                step.status === 'skipped' ? "border-zinc-500 text-zinc-500" :
                                "border-zinc-700 text-zinc-700"
                            )}>
                                {step.status === 'completed' ? <CheckCircle2 className="w-3 h-3" /> : 
                                 step.status === 'running' ? <div className="w-2 h-2 rounded-full bg-blue-500 animate-pulse" /> :
                                 <span className="text-[10px]">{i + 1}</span>}
                            </div>
                            {/* Only show label for active/surrounding steps to save space if needed, or truncate */}
                            <span className={cn(
                                "text-[10px] font-medium absolute top-8 whitespace-nowrap transition-colors max-w-[60px] truncate",
                                step.status === 'pending' ? "text-zinc-600" : "text-zinc-300"
                            )} title={step.name}>{step.name}</span>
                        </div>
                    ))}
                </div>
            </div>

            {/* Logs Area */}
            <ScrollArea className="flex-1 bg-black/40 p-2">
                <div className="space-y-2 p-2">
                    {steps.map((step) => (
                        <div key={step.id} className="rounded-lg border border-white/5 bg-white/[0.02] overflow-hidden">
                            <div 
                                className={cn(
                                    "flex items-center justify-between px-3 py-2 cursor-pointer hover:bg-white/5 transition-colors",
                                    step.status === 'running' && "bg-blue-500/5 border-l-2 border-l-blue-500"
                                )}
                                onClick={() => setExpandedStep(expandedStep === step.id ? null : step.id)}
                            >
                                <div className="flex items-center gap-2">
                                    {step.status === 'completed' ? <CheckCircle2 className="w-4 h-4 text-green-500" /> :
                                     step.status === 'running' ? <Clock className="w-4 h-4 text-blue-500 animate-spin-slow" /> :
                                     step.status === 'skipped' ? <Circle className="w-4 h-4 text-zinc-500" /> :
                                     <Circle className="w-4 h-4 text-zinc-700" />}
                                    <span className="text-sm font-medium">{step.name}</span>
                                    {step.status === 'completed' && <span className="text-[10px] px-1.5 py-0.5 bg-green-500/10 text-green-500 rounded font-mono">COMPLETED</span>}
                                    {step.status === 'skipped' && <span className="text-[10px] px-1.5 py-0.5 bg-zinc-500/10 text-zinc-500 rounded font-mono">SKIPPED</span>}
                                    {step.status === 'failed' && <span className="text-[10px] px-1.5 py-0.5 bg-red-500/10 text-red-500 rounded font-mono">FAILED</span>}
                                </div>
                                <div className="flex items-center gap-3">
                                    {step.duration && <span className="text-xs text-zinc-500 font-mono">{step.duration}</span>}
                                    {expandedStep === step.id ? <ChevronDown className="w-4 h-4 text-zinc-500" /> : <ChevronRight className="w-4 h-4 text-zinc-500" />}
                                </div>
                            </div>
                            
                            {/* Expanded Content */}
                            {expandedStep === step.id && (
                                <div className="border-t border-white/5 bg-black/20 p-3 space-y-3">
                                    {step.output ? (
                                        <div className="space-y-1">
                                            <div className="flex justify-between">
                                                <span className="text-xs text-zinc-500 uppercase tracking-wider">Output Payload</span>
                                            </div>
                                            <pre className="text-xs text-zinc-300 font-mono bg-zinc-900/50 p-2 rounded overflow-x-auto border border-white/5">
                                                {step.output}
                                            </pre>
                                        </div>
                                    ) : null}
                                    
                                    {(step.logs.length > 0) && (
                                         <div className="space-y-1">
                                             <span className="text-xs text-zinc-500 uppercase tracking-wider">Console Logs</span>
                                             {step.logs.map((log, i) => (
                                                <div key={i} className="text-xs text-zinc-400 font-mono pl-2 border-l border-zinc-700 break-all whitespace-pre-wrap">
                                                    {log}
                                                </div>
                                             ))}
                                        </div>
                                    )}
                                    
                                    {!step.output && step.logs.length === 0 && (
                                        <div className="text-xs text-zinc-600 italic">No output or logs available</div>
                                    )}
                                </div>
                            )}
                        </div>
                    ))}
                    
                    {steps.length === 0 && (
                        <div className="text-center text-zinc-500 py-8">
                            Initializing workflow nodes...
                        </div>
                    )}
                </div>
            </ScrollArea>
        </div>
      )}
    </div>
  );
}
