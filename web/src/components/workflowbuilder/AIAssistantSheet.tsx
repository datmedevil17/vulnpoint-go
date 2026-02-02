import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Textarea } from "@/components/ui/textarea";
import { Sparkles, Loader2, Bot, Wand2, Shield, Zap, Lock, ArrowRight } from "lucide-react";


interface AIAssistantSheetProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onGenerate: (prompt: string) => Promise<void>;
  isGenerating: boolean;
}

const AIAssistantSheet = ({
  open,
  onOpenChange,
  onGenerate,
  isGenerating,
}: AIAssistantSheetProps) => {
  const [prompt, setPrompt] = useState("");

  const handleGenerate = async () => {
    if (!prompt.trim()) return;
    await onGenerate(prompt);
  };

  const handleQuickPrompt = (text: string) => {
    setPrompt(text);
  };

  const templates = [
    {
      id: 1,
      icon: <Lock className="w-4 h-4 text-orange-500" />,
      title: "Secrets & SAST Pipeline",
      desc: "Gitleaks + Semgrep → GitHub Issue",
      prompt: "Scan my github repo for secrets and vulnerable configurations using Semgrep, then create a GitHub issue for findings."
    },
    {
      id: 2,
      icon: <Shield className="w-4 h-4 text-blue-500" />,
      title: "Daily Web Audit",
      desc: "Scheduled Nikto/OWASP → Email Report",
      prompt: "Run a full OWASP scan on my website every day at midnight. If vulnerabilities are found, email security@company.com."
    },
    {
      id: 3,
      icon: <Zap className="w-4 h-4 text-yellow-500" />,
      title: "Auto-Patch Dependencies",
      desc: "Trivy Scan → Auto-Fix PR for Criticals",
      prompt: "Check for new CVEs in my dependencies using Trivy. If critical, use auto-fix to create a PR updating the package."
    }
  ];

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="sm:max-w-lg flex flex-col h-full border-l border-border bg-background/95 backdrop-blur-xl">
        <SheetHeader className="pb-6 border-b border-border/50">
          <SheetTitle className="flex items-center gap-3 text-2xl font-bold bg-gradient-to-r from-indigo-500 to-purple-500 bg-clip-text text-transparent">
            <Bot className="w-8 h-8 text-indigo-500" />
            AI Architect
          </SheetTitle>
          <SheetDescription className="text-base text-muted-foreground">
            Describe your security goals, and I'll build the perfect workflow.
          </SheetDescription>
        </SheetHeader>

        <div className="flex-1 py-6 flex flex-col gap-6 overflow-hidden">
          {/* Quick Start Templates */}
          <div className="space-y-3 shrink-0">
             <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground px-1">
                <Wand2 className="w-4 h-4" />
                <span>Quick Start Templates</span>
             </div>
             <div className="grid gap-3">
                {templates.map((t) => (
                  <button 
                    key={t.id}
                    onClick={() => handleQuickPrompt(t.prompt)}
                    className="group relative flex items-center gap-4 p-3 rounded-xl border border-border/50 bg-card/50 hover:bg-accent hover:border-indigo-500/30 transition-all duration-300 text-left"
                  >
                    <div className="p-2 rounded-lg bg-background shadow-sm group-hover:scale-110 transition-transform duration-300">
                      {t.icon}
                    </div>
                    <div className="flex-1">
                      <div className="font-semibold text-sm text-foreground mb-0.5">{t.title}</div>
                      <div className="text-xs text-muted-foreground">{t.desc}</div>
                    </div>
                    <ArrowRight className="w-4 h-4 text-muted-foreground opacity-0 group-hover:opacity-100 -translate-x-2 group-hover:translate-x-0 transition-all" />
                  </button>
                ))}
             </div>
          </div>

          <div className="relative flex-1 flex flex-col">
            <div className={`
              absolute inset-0 rounded-2xl p-[1px] bg-gradient-to-b from-indigo-500/20 to-purple-500/20 pointer-events-none 
              ${Boolean(prompt) ? 'opacity-100' : 'opacity-50'} transition-opacity
            `} />
            <Textarea
              placeholder="Or describe your custom workflow needs..."
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              className="flex-1 resize-none border-none bg-card/30 p-6 text-base leading-relaxed focus-visible:ring-0 rounded-2xl"
            />
            {/* Character count or helper */}
            <div className="absolute bottom-4 right-4 text-xs text-muted-foreground/50 pointer-events-none">
              AI-Powered Generation
            </div>
          </div>
        </div>

        <SheetFooter className="border-t border-border/50 pt-6">
          <Button
            onClick={handleGenerate}
            disabled={!prompt.trim() || isGenerating}
            className="w-full h-12 text-base font-medium bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 shadow-lg shadow-indigo-500/20 transition-all duration-300 rounded-xl"
          >
            {isGenerating ? (
              <>
                <Loader2 className="w-5 h-5 mr-2 animate-spin" />
                Architecting Solution...
              </>
            ) : (
              <>
                <Sparkles className="w-5 h-5 mr-2" />
                Generate Workflow
              </>
            )}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
};

export default AIAssistantSheet;
