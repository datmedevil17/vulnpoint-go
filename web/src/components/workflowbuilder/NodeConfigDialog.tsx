import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { NodeType } from "@/types/workflow";

interface NodeConfigDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  nodeType: NodeType;
  initialData?: any;
  onSave: (data: any) => void;
}

const NodeConfigDialog = ({
  open,
  onOpenChange,
  nodeType,
  initialData,
  onSave,
}: NodeConfigDialogProps) => {
  // Email config
  const [email, setEmail] = useState(initialData?.email || "");

  // GitHub config
  const [repo, setRepo] = useState(initialData?.repo || "");

  // Slack config
  const [channel, setChannel] = useState(initialData?.channel || "");

  // GitHub repos
  const githubRepos: string[] = JSON.parse(
    localStorage.getItem("repos") || "[]"
  );

  const handleSave = () => {
    let configData = {};

    if (nodeType === "email") {
      configData = { email };
    } else if (nodeType === "github-issue") {
      configData = { repo };
    } else if (nodeType === "slack") {
      configData = { channel };
    } else if (nodeType === "auto-fix") {
      configData = {
        repo,
        path: initialData?.path || "",
        vulnerability: initialData?.vulnerability || ""
      };
    }

    onSave(configData);
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Configure {nodeType.replace("-", " ")} Node</DialogTitle>
          <DialogDescription>
            {nodeType === "email" && "Set the email address for notifications."}
            {nodeType === "github-issue" &&
              "Select the GitHub repository for issue creation."}
            {nodeType === "slack" &&
              "Configure the Slack channel for notifications."}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {nodeType === "email" && (
            <div className="space-y-2">
              <Label htmlFor="email">Email Recipient</Label>
              <Input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="recipient@example.com"
              />
              <p className="text-xs text-muted-foreground">
                📧 Security reports will be sent to this email address
              </p>
            </div>
          )}

          {nodeType === "github-issue" && (
            <div className="space-y-2">
              <Label htmlFor="repo">GitHub Repository</Label>
              <Select value={repo} onValueChange={setRepo}>
                <SelectTrigger>
                  <SelectValue placeholder="Select a repository" />
                </SelectTrigger>
                <SelectContent>
                  {githubRepos.map((repoName) => (
                    <SelectItem key={repoName} value={repoName}>
                      {repoName}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          {nodeType === "slack" && (
            <div className="space-y-2">
              <Label htmlFor="channel">Slack Channel</Label>
              <Input
                id="channel"
                value={channel}
                onChange={(e) => setChannel(e.target.value)}
                placeholder="#general"
              />
            </div>
          )}

          {nodeType === "estimate-cost" && (
            <div className="space-y-2">
              <Label htmlFor="budget">Monthly Budget Limit ($)</Label>
              <Input
                id="budget"
                type="number"
                defaultValue={initialData?.budget || "100"}
                onChange={(e) => {
                  initialData.budget = e.target.value;
                }}
                placeholder="1000"
              />
              <p className="text-xs text-muted-foreground">
                ⚠️ Workflow will require manual approval if estimated cost exceeds this amount.
              </p>
            </div>
          )}

          {nodeType === "policy-check" && (
            <div className="space-y-2">
              <Label htmlFor="policy">Compliance Standard</Label>
              <Select
                defaultValue={initialData?.policy || "cis"}
                onValueChange={(val) => (initialData.policy = val)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select a standard" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="cis">CIS Benchmarks Level 1</SelectItem>
                  <SelectItem value="cis-2">CIS Benchmarks Level 2</SelectItem>
                  <SelectItem value="soc2">SOC 2 Compliance</SelectItem>
                  <SelectItem value="gdpr">GDPR Privacy</SelectItem>
                  <SelectItem value="pci">PCI-DSS v4</SelectItem>
                </SelectContent>
              </Select>
            </div>
          )}

          {nodeType === "decision" && (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="variable">Variable to Check</Label>
                <Select
                  defaultValue={initialData?.variable || "vulnerabilities"}
                  onValueChange={(val) => (initialData.variable = val)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select variable" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="vulnerabilities">Vulnerability Count</SelectItem>
                    <SelectItem value="risk_score">Security Risk Score</SelectItem>
                    <SelectItem value="cost">Estimated Cost</SelectItem>
                    <SelectItem value="manual_input">User Input (Yes/No)</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              
              <div className="grid grid-cols-2 gap-4">
                 <div className="space-y-2">
                    <Label htmlFor="operator">Operator</Label>
                    <Select
                      defaultValue={initialData?.operator || "gt"}
                      onValueChange={(val) => (initialData.operator = val)}
                    >
                      <SelectTrigger>
                         <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="gt">{">"} Greater Than</SelectItem>
                        <SelectItem value="lt">{"<"} Less Than</SelectItem>
                        <SelectItem value="eq">{"="} Equals</SelectItem>
                        <SelectItem value="neq">{"!="} Not Equals</SelectItem>
                      </SelectContent>
                    </Select>
                 </div>
                 <div className="space-y-2">
                    <Label htmlFor="value">Threshold Value</Label>
                    <Input 
                      defaultValue={initialData?.value || "0"} 
                      onChange={(e) => (initialData.value = e.target.value)}
                    />
                 </div>
              </div>
            </div>
          )}

          {nodeType === "auto-fix" && (
            <>
              <div className="space-y-2">
                <Label htmlFor="repo">Override Repository (Optional)</Label>
                <Select value={repo} onValueChange={setRepo}>
                  <SelectTrigger>
                    <SelectValue placeholder="Inherit from Trigger" />
                  </SelectTrigger>
                  <SelectContent>
                    {githubRepos.map((repoName) => (
                      <SelectItem key={repoName} value={repoName}>
                        {repoName}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                   Defaults to the repository defined in the Trigger.
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="path">File Path (Optional)</Label>
                <Input
                  id="path"
                  defaultValue={initialData?.path || ""}
                  onChange={(e) => {
                     initialData.path = e.target.value;
                  }}
                  placeholder="e.g. main.go"
                />
                <p className="text-xs text-muted-foreground">
                   Leave blank to automatically fix files found by previous scanners (Dynamic).
                </p>
              </div>
            </>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSave}>Save</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default NodeConfigDialog;
