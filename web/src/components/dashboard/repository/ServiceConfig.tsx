import { useState } from "react";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Plus, Trash, Copy, Download, Code2, Server } from "lucide-react";
import { toast } from "sonner";

interface Service {
  id: string;
  name: string;
  path: string;
  language: string;
  hasDockerfile: boolean;
  port: string;
  database: string;
}

export function ServiceConfig() {
  const [services, setServices] = useState<Service[]>([
    { id: "1", name: "main", path: ".", language: "javascript", hasDockerfile: false, port: "3000", database: "mongodb" }
  ]);

  const addService = () => {
    setServices([...services, { 
      id: Math.random().toString(36).substr(2, 9), 
      name: "new-service", 
      path: ".", 
      language: "javascript", 
      hasDockerfile: false, 
      port: "", 
      database: "" 
    }]);
  };

  const removeService = (id: string) => {
    if (services.length > 1) {
      setServices(services.filter(s => s.id !== id));
    }
  };

  const updateService = (id: string, field: keyof Service, value: any) => {
    setServices(services.map(s => s.id === id ? { ...s, [field]: value } : s));
  };

  const generateJson = () => {
    const config: any = { services: {} };
    services.forEach(s => {
      config.services[s.name] = {
        path: s.path,
        has_dockerfile: s.hasDockerfile,
        language: s.language,
        framework: "react", // Mock inference
        ports: s.port ? [parseInt(s.port)] : [],
        databases: s.database ? [s.database] : []
      };
    });
    return JSON.stringify(config, null, 2);
  };

  const copyConfig = () => {
    navigator.clipboard.writeText(generateJson());
    toast.success("Configuration copied to clipboard");
  };

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
      {/* Left Column: Form */}
      <div className="space-y-6">
        <div className="flex items-center justify-between">
            <div>
                <h2 className="text-2xl font-bold tracking-tight">Configure Your Services</h2>
                <p className="text-muted-foreground">Manually specify your project's services and infrastructure.</p>
            </div>
        </div>

        <div className="space-y-4">
            <div className="flex justify-between items-center">
                <span className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Services</span>
                <Button variant="outline" size="sm" onClick={addService}>
                    <Plus className="w-4 h-4 mr-2" />
                    Add Service
                </Button>
            </div>

            {services.map((service) => (
                <Card key={service.id} className="border-l-4 border-l-primary/50">
                    <CardHeader className="py-3 px-4 bg-muted/20 flex flex-row items-center justify-between space-y-0">
                        <div className="flex items-center gap-2">
                             {service.language === 'javascript' || service.language === 'typescript' ? <Code2 className="w-4 h-4" /> : <Server className="w-4 h-4" />}
                             <span className="font-mono text-sm font-semibold">{service.name}</span>
                             <span className="text-xs px-2 py-0.5 rounded-full bg-background border uppercase">{service.language}</span>
                        </div>
                        <Button variant="ghost" size="icon" onClick={() => removeService(service.id)} disabled={services.length === 1}>
                            <Trash className="w-4 h-4 text-muted-foreground hover:text-destructive" />
                        </Button>
                    </CardHeader>
                    <CardContent className="p-4 space-y-4">
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label>Service Name</Label>
                                <Input value={service.name} onChange={(e) => updateService(service.id, 'name', e.target.value)} />
                            </div>
                            <div className="space-y-2">
                                <Label>Path</Label>
                                <Input value={service.path} onChange={(e) => updateService(service.id, 'path', e.target.value)} />
                            </div>
                        </div>

                        <div className="flex items-center space-x-2">
                            <Switch 
                                checked={service.hasDockerfile} 
                                onCheckedChange={(c: boolean) => updateService(service.id, 'hasDockerfile', c)} 
                            />
                            <Label>Has Dockerfile</Label>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                             <div className="space-y-2">
                                <Label>Language</Label>
                                <Select value={service.language} onValueChange={(v) => updateService(service.id, 'language', v)}>
                                    <SelectTrigger>
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="javascript">JavaScript</SelectItem>
                                        <SelectItem value="typescript">TypeScript</SelectItem>
                                        <SelectItem value="python">Python</SelectItem>
                                        <SelectItem value="go">Go</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                            <div className="space-y-2">
                                <Label>Port</Label>
                                <Input 
                                    placeholder="e.g. 3000" 
                                    value={service.port} 
                                    onChange={(e) => updateService(service.id, 'port', e.target.value)} 
                                />
                            </div>
                        </div>

                         <div className="space-y-2">
                                <Label>Databases</Label>
                                <div className="grid grid-cols-3 gap-2">
                                    {['postgresql', 'mongodb', 'redis', 'mysql'].map((db) => (
                                        <div 
                                            key={db}
                                            onClick={() => updateService(service.id, 'database', service.database === db ? '' : db)}
                                            className={`
                                                cursor-pointer flex items-center justify-center p-2 rounded-md border text-xs font-medium transition-all
                                                ${service.database === db ? 'bg-primary text-primary-foreground border-primary' : 'hover:bg-accent'}
                                            `}
                                        >
                                            {db}
                                        </div>
                                    ))}
                                </div>
                            </div>
                    </CardContent>
                </Card>
            ))}
        </div>
      </div>

      {/* Right Column: JSON Preview */}
      <div className="space-y-4">
        <div className="flex justify-between items-center">
             <span className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Generated JSON</span>
             <div className="flex gap-2">
                <Button variant="outline" size="sm" onClick={copyConfig}>
                    <Copy className="w-3.5 h-3.5 mr-2" />
                    Copy
                </Button>
                <Button variant="outline" size="sm">
                    <Download className="w-3.5 h-3.5 mr-2" />
                    Download
                </Button>
             </div>
        </div>
        <Card className="bg-zinc-950 text-zinc-50 border-zinc-800 h-full max-h-[600px] overflow-hidden flex flex-col">
            <div className="p-4 border-b border-zinc-800 bg-zinc-900/50 flex items-center gap-2">
                 <div className="w-3 h-3 rounded-full bg-red-500/20 border border-red-500/50" />
                 <div className="w-3 h-3 rounded-full bg-yellow-500/20 border border-yellow-500/50" />
                 <div className="w-3 h-3 rounded-full bg-green-500/20 border border-green-500/50" />
                 <span className="ml-2 text-xs font-mono text-zinc-500">service_profile.json</span>
            </div>
            <CardContent className="p-0 overflow-auto flex-1">
                <pre className="p-4 text-xs font-mono leading-relaxed text-green-400">
                    {generateJson()}
                </pre>
            </CardContent>
            
            <div className="p-4 bg-zinc-900 border-t border-zinc-800 text-xs text-zinc-500 space-y-1">
                <p className="font-semibold text-zinc-400">How to use:</p>
                <ol className="list-decimal list-inside space-y-1 ml-1">
                    <li>Download the <span className="text-zinc-300">service_profile.json</span> file</li>
                    <li>Place inside <span className="text-zinc-300">.infoundry/service_profile.json</span> in repo root</li>
                    <li>Run the pipeline - it will use your configuration!</li>
                </ol>
            </div>
        </Card>
      </div>
    </div>
  );
}
