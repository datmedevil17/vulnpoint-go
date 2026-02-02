import { useState, useCallback } from "react";
import api, { getErrorMessage } from "@/lib/api";
import { API_ENDPOINTS } from "@/lib/apiEndpoints";



interface ScanResult {
  id: string;
  type: string;
  target: string;
  status: string;
  result?: string;
  createdAt: string;
  completedAt?: string;
}

export const useScan = () => {
  const [scanning, setScanning] = useState(false);
  const [scanResults, setScanResults] = useState<ScanResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Run Nmap scan
  const runNmapScan = useCallback(async (target: string, ports: string = "1-1000") => {
    setScanning(true);
    setError(null);
    try {
      const response = await api.post<ScanResult>(API_ENDPOINTS.SCAN.NMAP, {
        target,
        ports,
      });
      return response.data;
    } catch (err) {
      const errorMessage = getErrorMessage(err);
      setError(errorMessage);
      throw err;
    } finally {
      setScanning(false);
    }
  }, []);

  // Run Nikto scan
  const runNiktoScan = useCallback(async (target: string) => {
    setScanning(true);
    setError(null);
    try {
      const response = await api.post<ScanResult>(API_ENDPOINTS.SCAN.NIKTO, {
        target,
      });
      return response.data;
    } catch (err) {
      const errorMessage = getErrorMessage(err);
      setError(errorMessage);
      throw err;
    } finally {
      setScanning(false);
    }
  }, []);

  // Run Gobuster scan
  const runGobusterScan = useCallback(async (target: string, wordlist: string) => {
    setScanning(true);
    setError(null);
    try {
      const response = await api.post<ScanResult>(API_ENDPOINTS.SCAN.GOBUSTER, {
        target,
        wordlist,
      });
      return response.data;
    } catch (err) {
      const errorMessage = getErrorMessage(err);
      setError(errorMessage);
      throw err;
    } finally {
      setScanning(false);
    }
  }, []);

  // Get all scan results
  const getScanResults = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await api.get<ScanResult[]>(API_ENDPOINTS.SCAN.RESULTS);
      setScanResults(response.data);
      return response.data;
    } catch (err) {
      const errorMessage = getErrorMessage(err);
      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // Get specific scan result
  const getScanResult = useCallback(async (id: string) => {
    setLoading(true);
    setError(null);
    try {
      const response = await api.get<ScanResult>(API_ENDPOINTS.SCAN.RESULT_BY_ID(id));
      return response.data;
    } catch (err) {
      const errorMessage = getErrorMessage(err);
      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // Clear error
  const clearError = useCallback(() => {
    setError(null);
  }, []);

  return {
    scanning,
    scanResults,
    loading,
    error,
    runNmapScan,
    runNiktoScan,
    runGobusterScan,
    getScanResults,
    getScanResult,
    clearError,
  };
};
