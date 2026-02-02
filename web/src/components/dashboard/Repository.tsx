import useGitHub from "@/hooks/useGithub";
import { useEffect, useMemo, useState } from "react";
import { RepositoryHeader } from "@/components/dashboard/repository/RepositoryHeader";
import { RepositoryCard } from "@/components/dashboard/repository/RepositoryCard";
import { RepositoryLoader } from "@/components/dashboard/repository/RepositoryLoader";
import { RepositoryError } from "@/components/dashboard/repository/RepositoryError";
import { Button } from "@/components/ui/button";
import { ChevronLeft, ChevronRight } from "lucide-react";

import { ServiceConfig } from "@/components/dashboard/repository/ServiceConfig";

const ITEMS_PER_PAGE = 6;

const RepositoryList = () => {
  const { repos, fetchRepositories, loading, error, clearError } = useGitHub();
  const [searchQuery, setSearchQuery] = useState("");
  const [currentPage, setCurrentPage] = useState(1);

  useEffect(() => {
    fetchRepositories();
  }, [fetchRepositories]);

  // Clear error when search query changes
  useEffect(() => {
    if (error) {
      clearError();
    }
  }, [searchQuery, clearError]);

  const filteredRepos = useMemo(() => {
    if (!searchQuery.trim()) return repos;

    const query = searchQuery.toLowerCase().trim();
    return repos.filter((repo) => {
      return (
        repo.name.toLowerCase().includes(query) ||
        (repo.language && repo.language.toLowerCase().includes(query)) ||
        (repo.private ? "private" : "public").includes(query) ||
        (repo.description && repo.description.toLowerCase().includes(query))
      );
    });
  }, [repos, searchQuery]);

  // Pagination Logic
  const totalPages = Math.ceil(filteredRepos.length / ITEMS_PER_PAGE);
  const paginatedRepos = useMemo(() => {
    const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
    return filteredRepos.slice(startIndex, startIndex + ITEMS_PER_PAGE);
  }, [filteredRepos, currentPage]);

  // Reset to page 1 when search changes
  useEffect(() => {
    setCurrentPage(1);
  }, [searchQuery]);

  const handleRetry = () => {
    clearError();
    fetchRepositories();
  };

  const handleSearch = (query: string) => {
    setSearchQuery(query);
  };

  const handlePreviousPage = () => {
    setCurrentPage((prev) => Math.max(prev - 1, 1));
  };

  const handleNextPage = () => {
    setCurrentPage((prev) => Math.min(prev + 1, totalPages));
  };

  if (error) {
    return <RepositoryError error={error} onRetry={handleRetry} />;
  }

  return (
    <div className="flex-1 px-4 py-6 sm:px-6 lg:px-8 flex flex-col h-full bg-white dark:bg-zinc-950 rounded-xl overflow-auto">
      <ServiceConfig />
      <RepositoryHeader
        totalRepos={filteredRepos.length}
        onSearch={handleSearch}
        onRefresh={handleRetry}
        loading={loading}
      />
      <div className="mt-2 divide-y divide-gray-200 dark:divide-gray-800 scrollbar-hidden overflow-scroll overflow-y-auto overflow-x-hidden flex-1">
        {loading ? (
          <RepositoryLoader />
        ) : filteredRepos.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <div className="text-gray-500 dark:text-gray-400 mb-4">
              {searchQuery ? (
                <>
                  <p className="text-lg font-medium mb-2">No repositories found</p>
                  <p className="text-sm">Try adjusting your search terms</p>
                </>
              ) : (
                <>
                  <p className="text-lg font-medium mb-2">No repositories available</p>
                  <p className="text-sm">Connect your GitHub account to get started</p>
                </>
              )}
            </div>
          </div>
        ) : (
          paginatedRepos.map((repo) => (
            <RepositoryCard key={`${repo.owner}-${repo.name}`} repo={repo} />
          ))
        )}
      </div>
      
      {/* Pagination Controls */}
      {filteredRepos.length > ITEMS_PER_PAGE && (
        <div className="py-4 border-t border-gray-200 dark:border-gray-800 flex items-center justify-between">
          <Button
            variant="outline"
            size="sm"
            onClick={handlePreviousPage}
            disabled={currentPage === 1}
          >
            <ChevronLeft className="h-4 w-4 mr-2" />
            Previous
          </Button>
          <span className="text-sm text-gray-500 dark:text-gray-400">
            Page {currentPage} of {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            onClick={handleNextPage}
            disabled={currentPage === totalPages}
          >
            Next
            <ChevronRight className="h-4 w-4 ml-2" />
          </Button>
        </div>
      )}
    </div>
  );
};

export default RepositoryList;
