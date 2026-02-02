import { Routes, Route, Navigate } from "react-router-dom";
import Auth from "@/pages/Auth";
import Dashboard from "@/pages/Dashboard";
import Workflow from "@/components/dashboard/Workflow";
import RepositoryList from "@/components/dashboard/Repository";
import Report from "@/components/dashboard/Report";
import PrivateRoute from "@/components/auth/PrivateRoute";
import WorkflowBuilder from "@/pages/WorkflowBuilder";
import AuthCallback from "@/pages/AuthCallback";
import Overview from "@/components/dashboard/Overview";


const AppRoutes = () => {
  return (
    <Routes>
      <Route path="/" element={<Auth />} />
      <Route path="/auth/callback" element={<AuthCallback />} />
      <Route element={<PrivateRoute />}>
        <Route path="/dashboard/*" element={<Dashboard />}>
          <Route index element={<Overview />} />
          <Route path="workflow" element={<Workflow />} />
          <Route path="repository" element={<RepositoryList />} />
          <Route path="report" element={<Report />} />
        </Route>
        <Route path="/workflow/:id" element={<WorkflowBuilder />} />
      </Route>
      <Route path="*" element={<Navigate to="/" />} />
    </Routes>
  );
};

export default AppRoutes;
