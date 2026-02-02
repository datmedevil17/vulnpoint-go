import { cn } from "@/lib/utils";
import { FileDiff } from "lucide-react";

export interface DiffChange {
  path: string;
  type: "create" | "update" | "delete" | "no-change";
  before?: string;
  after?: string;
  diff?: string; // Optional raw diff content
}

export interface DiffViewerProps {
  changes: DiffChange[];
  title?: string;
}

const DiffViewer: React.FC<DiffViewerProps> = ({ changes, title = "Infrastructure Changes" }) => {
  if (!changes || changes.length === 0) {
    return (
      <div className="border rounded-lg p-8 text-center bg-gray-50 dark:bg-zinc-900/50">
        <FileDiff className="w-10 h-10 mx-auto text-gray-400 mb-2" />
        <p className="text-muted-foreground">No changes detected.</p>
      </div>
    );
  }

  return (
    <div className="border rounded-lg overflow-hidden bg-background">
      {title && (
        <div className="bg-muted/50 px-4 py-3 border-b flex items-center justify-between">
          <h3 className="font-medium text-sm flex items-center gap-2">
            <FileDiff className="w-4 h-4" />
            {title}
          </h3>
          <span className="text-xs text-muted-foreground">
            {changes.filter(c => c.type !== 'no-change').length} resources changed
          </span>
        </div>
      )}

      <div className="divide-y">
        {changes.map((change, idx) => (
          <div key={idx} className="flex flex-col">
            {/* Header */}
            <div className="flex items-center gap-3 px-4 py-2 bg-muted/20">
              <span className={cn(
                "w-2 h-2 rounded-full",
                change.type === "create" && "bg-green-500",
                change.type === "delete" && "bg-red-500",
                change.type === "update" && "bg-yellow-500",
                change.type === "no-change" && "bg-gray-400"
              )} />
              <span className="font-mono text-sm font-medium flex-1">{change.path}</span>
              <span className={cn(
                "text-xs px-2 py-0.5 rounded-full uppercase font-bold",
                change.type === "create" && "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400",
                change.type === "delete" && "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400",
                change.type === "update" && "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400"
              )}>
                {change.type}
              </span>
            </div>

            {/* Diff Content */}
            <div className="p-0 overflow-x-auto text-xs font-mono">
              {change.type === "create" && change.after && (
                <pre className="p-4 bg-green-50/50 dark:bg-green-950/10 text-green-700 dark:text-green-400">
                  {`+ ${change.after.replace(/\n/g, "\n+ ")}`}
                </pre>
              )}
              {change.type === "delete" && change.before && (
                <pre className="p-4 bg-red-50/50 dark:bg-red-950/10 text-red-700 dark:text-red-400">
                  {`- ${change.before.replace(/\n/g, "\n- ")}`}
                </pre>
              )}
              {change.type === "update" && (
                <div className="grid grid-cols-2 divide-x">
                  <div className="p-4 bg-red-50/20 dark:bg-red-950/5 text-red-600 dark:text-red-400 opacity-80">
                    <div className="mb-2 text-[10px] uppercase tracking-wider text-muted-foreground">Before</div>
                    <pre>{change.before || "(empty)"}</pre>
                  </div>
                  <div className="p-4 bg-green-50/20 dark:bg-green-950/5 text-green-600 dark:text-green-400">
                     <div className="mb-2 text-[10px] uppercase tracking-wider text-muted-foreground">After</div>
                    <pre>{change.after || "(empty)"}</pre>
                  </div>
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default DiffViewer;
