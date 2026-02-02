import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { 
  Activity, 
  ShieldAlert, 
  CheckCircle, 
  XCircle, 
  BarChart3, 
  ArrowUpRight, 
  ArrowDownRight,
  Plus,
  Zap
} from "lucide-react";
import { useNavigate } from "react-router-dom";

const Overview = () => {
  const navigate = useNavigate();

  // Mock Data for "Command Center"
  const stats = [
    {
      title: "Total Scans",
      value: "1,248",
      change: "+12%",
      trend: "up",
      icon: Activity,
      color: "text-blue-500",
      bg: "bg-blue-500/10",
    },
    {
      title: "Critical Risks",
      value: "3",
      change: "-2",
      trend: "down",
      icon: ShieldAlert,
      color: "text-red-500",
      bg: "bg-red-500/10",
    },
    {
      title: "Compliance Score",
      value: "94%",
      change: "+2%",
      trend: "up",
      icon: CheckCircle,
      color: "text-green-500",
      bg: "bg-green-500/10",
    },
    {
      title: "Failed Builds",
      value: "12",
      change: "+4",
      trend: "up", // bad trend
      icon: XCircle,
      color: "text-orange-500",
      bg: "bg-orange-500/10",
    },
  ];

  const recentActivity = [
    { id: 1, action: "Production Deployment", user: "AI Auto-Fix", time: "2 mins ago", status: "success" },
    { id: 2, action: "Drift Detected (AWS SG)", user: "System", time: "15 mins ago", status: "warning" },
    { id: 3, action: "Compliance Audit", user: "j.doe@example.com", time: "1 hour ago", status: "success" },
    { id: 4, action: "Failed Cost Estimate", user: "System", time: "2 hours ago", status: "failed" },
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Command Center</h2>
          <p className="text-muted-foreground">Real-time overview of your security infrastructure.</p>
        </div>
        <div className="flex gap-2">
          <Button onClick={() => navigate("/dashboard/workflow/new")}>
            <Plus className="mr-2 h-4 w-4" /> New Workflow
          </Button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat, i) => (
          <Card key={i} className="border-none shadow-sm bg-card/50 backdrop-blur-sm hover:bg-card/80 transition-colors">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {stat.title}
              </CardTitle>
              <div className={`p-2 rounded-full ${stat.bg}`}>
                <stat.icon className={`h-4 w-4 ${stat.color}`} />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stat.value}</div>
              <p className="text-xs text-muted-foreground flex items-center mt-1">
                {stat.trend === "up" ? (
                  <ArrowUpRight className={`h-3 w-3 mr-1 ${stat.title === "Failed Builds" ? "text-red-500" : "text-green-500"}`} />
                ) : (
                  <ArrowDownRight className="h-3 w-3 mr-1 text-green-500" />
                )}
                <span className={stat.trend === "up" ? "text-green-500" : "text-green-500"}>
                  {stat.change}
                </span>
                <span className="ml-1 text-muted-foreground">from last month</span>
              </p>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
        {/* Main Chart Area (Mock) */}
        <Card className="col-span-4 border-none shadow-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
               <BarChart3 className="h-5 w-5 text-gray-500" />
               Security Posture Trend
            </CardTitle>
          </CardHeader>
          <CardContent className="pl-2">
            <div className="h-[240px] flex items-end justify-between gap-2 px-4 pb-4 pt-10">
              {[65, 59, 80, 81, 56, 55, 40, 70, 75, 85, 90, 95].map((h, i) => (
                <div key={i} className="w-full bg-primary/10 hover:bg-primary/20 rounded-t-sm relative group transition-all" style={{ height: `${h}%` }}>
                    <div className="absolute -top-8 left-1/2 -translate-x-1/2 bg-popover text-popover-foreground text-xs py-1 px-2 rounded opacity-0 group-hover:opacity-100 transition-opacity">
                        {h}%
                    </div>
                </div>
              ))}
            </div>
            <div className="flex justify-between text-xs text-muted-foreground px-4">
                <span>Jan</span><span>Dec</span>
            </div>
          </CardContent>
        </Card>

        {/* Recent Activity */}
        <Card className="col-span-3 border-none shadow-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
                <Zap className="h-5 w-5 text-yellow-500" />
                Recent Activity
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-8">
              {recentActivity.map((activity) => (
                <div key={activity.id} className="flex items-center">
                  <span className={`relative flex h-2 w-2 mr-4 rounded-full ${
                      activity.status === 'success' ? 'bg-green-500' :
                      activity.status === 'warning' ? 'bg-yellow-500' :
                      'bg-red-500'
                  }`}>
                    <span className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${
                        activity.status === 'success' ? 'bg-green-400' :
                        activity.status === 'warning' ? 'bg-yellow-400' :
                        'bg-red-400'
                    }`}></span>
                  </span>
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">{activity.action}</p>
                    <p className="text-xs text-muted-foreground">
                      {activity.user} · {activity.time}
                    </p>
                  </div>
                  <div className="ml-auto font-medium text-xs text-muted-foreground">
                    {activity.status === 'success' ? 'Completed' : 
                     activity.status === 'warning' ? 'Review' : 'Failed'}
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default Overview;
