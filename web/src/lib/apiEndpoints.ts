/**
 * API Endpoints
 * 
 * Centralized API endpoint definitions to prevent endpoint mismatches
 * between frontend and backend.
 */

export const API_ENDPOINTS = {
  // Authentication
  AUTH: {
    GITHUB_URL: '/api/auth/github',
    GITHUB_CALLBACK: '/api/auth/github/callback',
    USER: '/api/user',
    LOGOUT: '/api/auth/logout',
  },

  // GitHub Integration
  GITHUB: {
    REPOSITORIES: '/api/github/repositories',
    REPOSITORY_FILES: (owner: string, repo: string) => 
      `/api/github/repositories/${owner}/${repo}/files`,
    FILE_CONTENT: (owner: string, repo: string) => 
      `/api/github/repositories/${owner}/${repo}/content`,
  },

  // Workflows
  WORKFLOWS: {
    BASE: '/api/workflows',
    BY_ID: (id: string) => `/api/workflows/${id}`,
    EXECUTE: (id: string) => `/api/workflows/${id}/execute`,
    STATUS: (id: string) => `/api/workflows/executions/${id}`,
    REPORTS: '/api/workflows/reports',
  },

  // Code Analysis
  CODE: {
    ANALYZE: '/api/code/analyze',
    QUICK_SCAN: '/api/code/quick-scan',
    COMPARE: '/api/code/compare',
  },

  // Chatbot
  CHATBOT: {
    CHAT: '/api/chatbot/chat',
    EXPLAIN: '/api/chatbot/explain',
    REMEDIATE: '/api/chatbot/remediate',
    ASK: '/api/chatbot/ask',
  },

  // Scanner
  SCAN: {
    NMAP: '/api/scan/nmap',
    NIKTO: '/api/scan/nikto',
    GOBUSTER: '/api/scan/gobuster',
    RESULTS: '/api/scan/results',
    RESULT_BY_ID: (id: string) => `/api/scan/results/${id}`,
  },

  // Health Check
  HEALTH: '/api/health',
} as const;
