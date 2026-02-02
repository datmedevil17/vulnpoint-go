import { Workflow } from "@/types/workflow";
import api from "@/lib/api";
import { API_ENDPOINTS } from "@/lib/apiEndpoints";

export const workflowApi = {
  getAllWorkflows: async (): Promise<Workflow[]> => {
    try {
      const response = await api.get<{ success: boolean; data: Workflow[] }>(
        API_ENDPOINTS.WORKFLOWS.BASE
      );
      return response.data.data || [];
    } catch (error) {
      console.error("Error fetching workflows:", error);
      throw error;
    }
  },

  getWorkflowById: async (id: string): Promise<Workflow> => {
    try {
      const response = await api.get<{ success: boolean; data: Workflow }>(
        API_ENDPOINTS.WORKFLOWS.BY_ID(id)
      );
      
      const workflowData = response.data.data;
      
      if (!workflowData || !workflowData.id) {
        throw new Error("Invalid response: workflow data missing or incomplete");
      }
      
      return workflowData;
    } catch (error) {
      console.error(`Error fetching workflow ${id}:`, error);
      throw error;
    }
  },

  createWorkflow: async (workflow: Workflow): Promise<Workflow> => {
    try {
      const response = await api.post<{ success: boolean; data: Workflow }>(
        API_ENDPOINTS.WORKFLOWS.BASE,
        workflow
      );
      
      const data = response.data;
      
      // Backend wraps response in { success: true, data: {...} }
      const workflowData = data.data;
      
      if (!workflowData) {
        throw new Error("Invalid response: workflow data missing");
      }
      
      if (!workflowData.id || workflowData.id === '') {
        throw new Error("Invalid response: workflow ID missing or empty");
      }
      
      return workflowData;
    } catch (error) {
      console.error("Error creating workflow:", error);
      throw error;
    }
  },

  updateWorkflow: async (workflow: Workflow): Promise<Workflow> => {
    try {
      const response = await api.put<{ success: boolean; data: Workflow }>(
        API_ENDPOINTS.WORKFLOWS.BY_ID(workflow.id),
        workflow
      );
      return response.data.data;
    } catch (error) {
      console.error(`Error updating workflow ${workflow.id}:`, error);
      throw error;
    }
  },

  deleteWorkflow: async (id: string): Promise<void> => {
    try {
      await api.delete(API_ENDPOINTS.WORKFLOWS.BY_ID(id));
    } catch (error) {
      console.error(`Error deleting workflow ${id}:`, error);
      throw error;
    }
  },

  executeWorkflow: async (id: string): Promise<any> => {
    try {
      const response = await api.post(API_ENDPOINTS.WORKFLOWS.EXECUTE(id));
      return response.data;
    } catch (error) {
      console.error(`Error executing workflow ${id}:`, error);
      throw error;
    }
  },

  getExecution: async (id: string): Promise<any> => {
    try {
      const response = await api.get(API_ENDPOINTS.WORKFLOWS.STATUS(id));
      return response.data;
    } catch (error) {
      console.error(`Error getting execution ${id}:`, error);
      throw error;
    }
  },

  getExecutionStatus: async (id: string): Promise<any> => {
    try {
      const response = await api.get(API_ENDPOINTS.WORKFLOWS.STATUS(id));
      return response.data;
    } catch (error) {
      console.error(`Error getting execution status ${id}:`, error);
      throw error;
    }
  },

  getAllExecutionResults: async (): Promise<any> => {
    try {
      const response = await api.get(API_ENDPOINTS.WORKFLOWS.REPORTS);
      return response.data;
    } catch (error) {
      console.error("Error getting execution results:", error);
      throw error;
    }
  },
  
  deleteExecutionResult: (id: string) => api.delete(`/workflows/reports/${id}`),

  executeNode: (nodeData: any) => api.post("/workflows/execute-node", nodeData),
  
  generateWorkflow: (prompt: string) => api.post("/workflow/ai-generate", { prompt }),
};
