import { useState, useEffect } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { 
  PlusCircle, 
  CheckCircle, 
  Clock, 
  XCircle, 
  Eye, 
  RefreshCw,
  AlertTriangle,
  Shield,
  Bug,
  Mail,
  MessageSquare,
  Github,
  FileText,
  Trash2
} from "lucide-react";
import { workflowApi } from "@/hooks/useWorkflow";
import { toast } from "sonner";
import DiffViewer from "@/components/shared/DiffViewer";

// Define types for execution results
interface ExecutionResult {
  id: string;
  workflowId: string;
  name: string;
  status: "completed" | "failed" | "running";
  startedAt: string;
  completedAt?: string;
  results?: Record<string, any>;
  error?: string;
  duration?: number;
}

interface NodeResult {
  type: string;
  success: boolean;
  data?: any;
  error?: string;
  timestamp: string;
}

const ReportCardPage = () => {
  const [reports, setReports] = useState<ExecutionResult[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedReport, setSelectedReport] = useState<ExecutionResult | null>(null);
  const [showDetails, setShowDetails] = useState(false);

  // Fetch execution results
  const fetchReports = async () => {
    try {
      setLoading(true);
      const response = await workflowApi.getAllExecutionResults();
      setReports(response.data || []);
    } catch (error) {
      console.error("Error fetching reports:", error);
      toast.error("Failed to load reports");
    } finally {
      setLoading(false);
    }
  };

  // Format relative time
  const getRelativeTime = (dateString: string) => {
    if (!dateString) return "Pending..."; // Handle missing date
    
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return "Invalid Date";

    const now = new Date();
    const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (diffInSeconds < 5) return "just now";
    if (diffInSeconds < 60) return `${diffInSeconds}s ago`;
    if (diffInSeconds < 120) return "1 min ago";
    if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)} mins ago`;
    if (diffInSeconds < 7200) return "1 hour ago";
    if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)} hours ago`;
    return date.toLocaleDateString();
  };

  // Format duration
  const formatDuration = (duration: number) => {
    if (duration < 1000) return `${duration}ms`;
    if (duration < 60000) return `${Math.floor(duration / 1000)}s`;
    return `${Math.floor(duration / 60000)}m ${Math.floor((duration % 60000) / 1000)}s`;
  };

  // Get status color and icon
  const getStatusInfo = (status: string) => {
    switch (status) {
      case "completed":
        return { 
          color: "bg-green-500/20 text-green-500", 
          icon: CheckCircle, 
          label: "Completed" 
        };
      case "failed":
        return { 
          color: "bg-red-500/20 text-red-500", 
          icon: XCircle, 
          label: "Failed" 
        };
      case "running":
        return { 
          color: "bg-blue-500/20 text-blue-500", 
          icon: Clock, 
          label: "Running" 
        };
      default:
        return { 
          color: "bg-gray-500/20 text-gray-500", 
          icon: Clock, 
          label: "Unknown" 
        };
    }
  };

  // Get node type icon
  const getNodeIcon = (type: string) => {
    switch (type) {
      case "nmap": return Shield;
      case "gobuster": return Bug;
      case "sqlmap": return AlertTriangle;
      case "wpscan": return Bug;
      case "email": return Mail;
      case "slack": return MessageSquare;
      case "github-issue": return Github;
      default: return AlertTriangle;
    }
  };

  // Re-execute workflow
  const handleReexecute = async (workflowId: string) => {
    try {
      await workflowApi.executeWorkflow(workflowId);
      toast.success("Workflow execution started");
      // Refresh reports after a delay
      setTimeout(fetchReports, 2000);
    } catch (error) {
      toast.error("Failed to execute workflow");
    }
  };

  // Delete workflow report
  const handleDelete = async (reportId: string) => {
      try {
          if (!window.confirm("Are you sure you want to delete this report?")) return;
          
          await workflowApi.deleteExecutionResult(reportId);
          toast.success("Report deleted successfully");
          // Update local state immediately
          setReports(prev => prev.filter(r => r.id !== reportId));
      } catch (error) {
          console.error("Delete failed", error);
          toast.error("Failed to delete report");
      }
  }

  // View report details
  const viewDetails = (report: ExecutionResult) => {
    setSelectedReport(report);
    setShowDetails(true);
  };

  // Format AI report for display
  const formatAIReport = (reportText: string) => {
    if (!reportText) return 'No report available';
    
    // Convert markdown-style headers to HTML with enhanced styling
    let formatted = reportText
      // Main headers with emoji support
      .replace(/^# (.*$)/gm, '<h1 class="text-2xl font-bold mt-6 mb-4 text-blue-700 dark:text-blue-400 border-b border-gray-200 dark:border-gray-700 pb-2">$1</h1>')
      .replace(/^## (.*$)/gm, '<h2 class="text-xl font-bold mt-5 mb-3 text-blue-600 dark:text-blue-400">$1</h2>')
      .replace(/^### (.*$)/gm, '<h3 class="text-lg font-semibold mt-4 mb-2 text-gray-800 dark:text-gray-200">$1</h3>')
      
      // Handle different types of emphasis
      .replace(/\*\*\*(.*?)\*\*\*/gm, '<strong class="font-bold text-orange-600 dark:text-orange-400">$1</strong>')
      .replace(/\*\*(.*?)\*\*/gm, '<strong class="font-semibold text-gray-900 dark:text-gray-100">$1</strong>')
      .replace(/\*(.*?)\*/gm, '<em class="italic">$1</em>')
      
      // Handle different types of alerts with better styling
      .replace(/🚨\s*\*\*URGENT\*\*/gm, '<span class="inline-flex items-center px-3 py-1 text-sm font-semibold text-red-700 bg-red-100 border border-red-300 rounded-full dark:text-red-400 dark:bg-red-900/30 dark:border-red-800">🚨 URGENT</span>')
      .replace(/⚠️\s*\*\*Important\*\*/gm, '<span class="inline-flex items-center px-3 py-1 text-sm font-semibold text-orange-700 bg-orange-100 border border-orange-300 rounded-full dark:text-orange-400 dark:bg-orange-900/30 dark:border-orange-800">⚠️ Important</span>')
      .replace(/📝\s*\*\*Good to Fix\*\*/gm, '<span class="inline-flex items-center px-3 py-1 text-sm font-semibold text-blue-700 bg-blue-100 border border-blue-300 rounded-full dark:text-blue-400 dark:bg-blue-900/30 dark:border-blue-800">📝 Good to Fix</span>')
      .replace(/✅\s*\*\*No Critical Issues Found\*\*/gm, '<span class="inline-flex items-center px-3 py-1 text-sm font-semibold text-green-700 bg-green-100 border border-green-300 rounded-full dark:text-green-400 dark:bg-green-900/30 dark:border-green-800">✅ No Critical Issues Found</span>')
      
      // Handle security grades with special styling
      .replace(/Your Security Grade:\s*([A-F])/gm, '<div class="my-4 p-4 bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 border border-blue-200 dark:border-blue-800 rounded-lg"><span class="text-lg font-bold">Your Security Grade: <span class="text-2xl font-bold text-blue-600 dark:text-blue-400">$1</span></span></div>')
      
      // Handle lists with better spacing
      .replace(/^- (.*$)/gm, '<li class="ml-4 mb-1 text-gray-700 dark:text-gray-300">$1</li>')
      .replace(/^\d+\.\s+(.*$)/gm, '<li class="ml-4 mb-1 text-gray-700 dark:text-gray-300">$1</li>')
      
      // Handle action items with special formatting
      .replace(/\*\*What this means:\*\*/gm, '<div class="mt-2 mb-1"><strong class="text-blue-600 dark:text-blue-400">What this means:</strong>')
      .replace(/\*\*Business risk:\*\*/gm, '<div class="mt-1 mb-1"><strong class="text-red-600 dark:text-red-400">Business risk:</strong>')
      .replace(/\*\*Action needed:\*\*/gm, '<div class="mt-1 mb-2"><strong class="text-green-600 dark:text-green-400">Action needed:</strong>')
      .replace(/\*\*What to do:\*\*/gm, '<div class="mt-2 mb-1"><strong class="text-purple-600 dark:text-purple-400">What to do:</strong>');
    
    // Wrap consecutive list items in proper ul/ol tags
    formatted = formatted.replace(/((<li class="ml-4 mb-1[^"]*">.*?<\/li>\s*)+)/g, '<ul class="mb-4 space-y-1">$1</ul>');
    
    // Add proper spacing for paragraphs and preserve line breaks
    formatted = formatted
      .replace(/\n\n/g, '</p><p class="mb-3 leading-relaxed text-gray-700 dark:text-gray-300">')
      .replace(/\n/g, '<br>');
    
    // Wrap in paragraph tags with better styling
    if (!formatted.startsWith('<h1') && !formatted.startsWith('<h2') && !formatted.startsWith('<div')) {
      formatted = '<p class="mb-3 leading-relaxed text-gray-700 dark:text-gray-300">' + formatted + '</p>';
    }
    
    return formatted;
  };

  useEffect(() => {
    fetchReports();
    // Auto-refresh every 30 seconds
    const interval = setInterval(fetchReports, 30000);
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="p-6">
        <div className="max-w-6xl mx-auto">
          <div className="flex items-center justify-center h-64">
            <RefreshCw className="h-8 w-8 animate-spin" />
            <span className="ml-2">Loading reports...</span>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="max-w-6xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-bold">Workflow Reports</h1>
          <Button onClick={fetchReports} variant="outline" size="sm">
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
        </div>

        {reports.length === 0 ? (
          <Card>
            <CardContent className="text-center py-12">
              <Clock className="h-12 w-12 mx-auto mb-4 text-gray-400" />
              <h3 className="text-lg font-medium mb-2">No Reports Yet</h3>
              <p className="text-gray-500">Execute some workflows to see reports here.</p>
            </CardContent>
          </Card>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {reports.map((report) => {
              const statusInfo = getStatusInfo(report.status);
              const StatusIcon = statusInfo.icon;
              
              return (
                <Card key={report.id} className="hover:shadow-md transition-shadow">
                  <CardContent className="p-4">
                    <div className="flex items-center justify-between mb-3">
                      <h3 className="font-medium truncate">{report.name}</h3>
                      <StatusIcon className={`h-4 w-4 ${
                        report.status === "completed" ? "text-green-500" :
                        report.status === "failed" ? "text-red-500" : "text-blue-500"
                      }`} />
                    </div>

                    <div className="space-y-2 mb-4">
                      <div className="text-xs text-gray-500">
                        {getRelativeTime(report.startedAt)}
                      </div>
                      
                      {report.duration && (
                        <div className="text-xs text-gray-500">
                          Duration: {formatDuration(report.duration)}
                        </div>
                      )}

                      <Badge className={`text-xs ${statusInfo.color}`}>
                        {statusInfo.label}
                      </Badge>
                    </div>

                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        className="flex-1"
                        onClick={() => viewDetails(report)}
                      >
                        <Eye className="h-3 w-3 mr-1" />
                        View
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleReexecute(report.workflowId)}
                      >
                        <PlusCircle className="h-3 w-3 mr-1" />
                        Run
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost" 
                        className="text-red-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/10"
                        onClick={(e) => {
                           e.stopPropagation();
                           handleDelete(report.id);
                        }}
                      >
                        <Trash2 className="h-3 w-3" />
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              );
            })}
          </div>
        )}

        {/* Details Modal */}
        {showDetails && selectedReport && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <div className="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-4xl w-full mx-4 max-h-[80vh] overflow-hidden">
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-xl font-bold">{selectedReport.name} - Results</h2>
                <Button variant="ghost" onClick={() => setShowDetails(false)}>×</Button>
              </div>
              
              <ScrollArea className="h-96">
                <div className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <strong>Status:</strong> {selectedReport.status}
                    </div>
                    <div>
                      <strong>Started:</strong> {new Date(selectedReport.startedAt).toLocaleString()}
                    </div>
                    {selectedReport.completedAt && (
                      <div>
                        <strong>Completed:</strong> {new Date(selectedReport.completedAt).toLocaleString()}
                      </div>
                    )}
                    {selectedReport.duration && (
                      <div>
                        <strong>Duration:</strong> {formatDuration(selectedReport.duration)}
                      </div>
                    )}
                  </div>

                  {selectedReport.error && (
                    <div className="bg-red-50 dark:bg-red-900/20 p-3 rounded">
                      <strong className="text-red-600">Error:</strong>
                      <pre className="text-sm mt-1 whitespace-pre-wrap">{selectedReport.error}</pre>
                    </div>
                  )}

                  {selectedReport.results && (
                    <div>
                      {/* AI Report Section */}
                      {selectedReport.results.ai_report && (
                        <div className="mb-8">
                          <div className="flex items-center gap-3 mb-6">
                            <div className="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
                              <FileText className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                            </div>
                            <div>
                              <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                                Your Website Security Report
                              </h3>
                              <p className="text-sm text-gray-600 dark:text-gray-400">
                                Easy-to-understand security analysis and recommendations
                              </p>
                            </div>
                          </div>
                          
                          <Card className="border-0 shadow-lg bg-gradient-to-br from-white to-gray-50 dark:from-gray-800 dark:to-gray-900">
                            <CardContent className="p-6">
                              {/* Security Grade Display */}
                              {selectedReport.results.ai_report.security_grade && (
                                <div className="mb-6 p-4 bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 border border-blue-200 dark:border-blue-800 rounded-lg text-center">
                                  <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Security Grade</div>
                                  <div className={`text-4xl font-bold ${
                                    selectedReport.results.ai_report.security_grade === 'A' ? 'text-green-600' :
                                    selectedReport.results.ai_report.security_grade === 'B' ? 'text-blue-600' :
                                    selectedReport.results.ai_report.security_grade === 'C' ? 'text-yellow-600' :
                                    selectedReport.results.ai_report.security_grade === 'D' ? 'text-orange-600' :
                                    'text-red-600'
                                  }`}>
                                    {selectedReport.results.ai_report.security_grade}
                                  </div>
                                  {selectedReport.results.ai_report.total_issues !== undefined && (
                                    <div className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                                      {selectedReport.results.ai_report.total_issues} issues found
                                      {selectedReport.results.ai_report.critical_issues > 0 && (
                                        <span className="ml-2 px-2 py-1 bg-red-100 text-red-700 rounded-full text-xs font-semibold">
                                          {selectedReport.results.ai_report.critical_issues} critical
                                        </span>
                                      )}
                                    </div>
                                  )}
                                </div>
                              )}
                              
                              <div className="prose dark:prose-invert max-w-none">
                                <div 
                                  className="text-sm leading-relaxed"
                                  dangerouslySetInnerHTML={{ 
                                    __html: formatAIReport(selectedReport.results.ai_report.ai_report) 
                                  }}
                                />
                              </div>
                              
                              <div className="mt-6 pt-4 border-t border-gray-200 dark:border-gray-700 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
                                <div className="flex items-center gap-2">
                                  <Shield className="h-3 w-3" />
                                  <span>Generated by: {selectedReport.results.ai_report.generated_by || 'VulnPilot AI'}</span>
                                </div>
                                <div className="flex items-center gap-2">
                                  <Clock className="h-3 w-3" />
                                  <span>{new Date(selectedReport.results.ai_report.report_date).toLocaleString()}</span>
                                </div>
                              </div>
                            </CardContent>
                          </Card>
                        </div>
                      )}

                      {selectedReport.results.ai_report_error && (
                        <div className="mb-6 p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
                          <div className="flex items-center gap-2 mb-2">
                            <AlertTriangle className="h-4 w-4 text-yellow-600" />
                            <strong className="text-yellow-700 dark:text-yellow-400">AI Report Generation Failed</strong>
                          </div>
                          <div className="text-sm text-yellow-600 dark:text-yellow-300">{selectedReport.results.ai_report_error}</div>
                        </div>
                      )}

                      {/* Raw Scan Results - Less prominent when AI report is available */}
                      <div className={`${selectedReport.results.ai_report ? 'mt-8' : 'mt-0'}`}>
                        <div className="flex items-center gap-2 mb-4">
                          <div className="p-1.5 bg-gray-100 dark:bg-gray-800 rounded">
                            <Bug className="h-4 w-4 text-gray-600 dark:text-gray-400" />
                          </div>
                          <h3 className="font-medium text-gray-700 dark:text-gray-300">Technical Scan Details</h3>
                          {selectedReport.results.ai_report && (
                            <Badge variant="outline" className="text-xs">
                              For technical review
                            </Badge>
                          )}
                        </div>
                        <div className="space-y-3">
                        {Object.entries(selectedReport.results).filter(([key]) => 
                          !key.startsWith('ai_report')).map(([nodeId, result]) => {
                          const nodeResult = result as NodeResult;
                          const NodeIcon = getNodeIcon(nodeResult.type);
                          
                          return (
                            <Card key={nodeId} className="p-3">
                              <div className="flex items-center gap-2 mb-2">
                                <NodeIcon className="h-4 w-4" />
                                <span className="font-medium capitalize">{nodeResult.type}</span>
                                <Badge variant={nodeResult.success ? "default" : "destructive"}>
                                  {nodeResult.success ? "Success" : "Failed"}
                                </Badge>
                              </div>
                              
                              {nodeResult.error && (
                                <div className="text-red-600 text-sm mb-2">{nodeResult.error}</div>
                              )}
                              
                              {nodeResult.data && (
                                <details className="cursor-pointer">
                                  <summary className="text-sm font-medium mb-2">
                                    View Raw Data ({Object.keys(nodeResult.data).length} items)
                                  </summary>
                                  <pre className="text-xs bg-gray-100 dark:bg-gray-700 p-2 rounded overflow-auto max-h-48">
                                    {JSON.stringify(nodeResult.data, null, 2)}
                                  </pre>
                                </details>
                              )}
                            </Card>
                          );
                        })}
                        </div>
                      </div>

                      {/* Infrastructure Diffs Section */}
                      {selectedReport.results && Object.entries(selectedReport.results).some(([, result]: any) => result && result.changes && result.changes.length > 0) && (
                        <div className="mt-8">
                          <div className="flex items-center gap-2 mb-4">
                            <div className="p-1.5 bg-purple-100 dark:bg-purple-900/30 rounded">
                              <FileText className="h-4 w-4 text-purple-600 dark:text-purple-400" />
                            </div>
                            <h3 className="font-medium text-gray-700 dark:text-gray-300">Infrastructure Changes</h3>
                          </div>
                          
                          <div className="space-y-6">
                            {Object.entries(selectedReport.results).map(([nodeId, result]: any) => {
                              if (!result || !result.changes || result.changes.length === 0) return null;
                              
                              return (
                                <div key={nodeId} className="space-y-2">
                                  <h4 className="text-sm font-medium flex items-center gap-2">
                                    <span className="text-muted-foreground">Node:</span> 
                                    {nodeId}
                                    <Badge variant="outline" className="text-xs font-normal">
                                      {result.type}
                                    </Badge>
                                  </h4>
                                  <DiffViewer changes={result.changes} />
                                </div>
                              );
                            })}
                          </div>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              </ScrollArea>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default ReportCardPage;
