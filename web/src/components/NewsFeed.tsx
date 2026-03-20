import React, { useEffect, useState } from "react";
import { api } from "@/lib/api";

interface Article {
  title: string;
  description: string;
  url: string;
  urlToImage: string;
  publishedAt: string;
  source: { name: string };
}

// Featured articles come from the pins table with a different shape
interface FeaturedPin {
  title: string;
  url: string;
  articleId: string;
  description: string;
  imageUrl: string;
  source: string;
  pinnedAt: string;
}

interface Props {
  onArticleClick: (article: Article) => void;
}

export default function NewsFeed({ onArticleClick }: Props) {
  const [articles, setArticles] = useState<Article[]>([]);
  const [searchInput, setSearchInput] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [isSearchResults, setIsSearchResults] = useState(false);

  const fetchFeatured = async () => {
    setLoading(true);
    setError("");
    try {
      const data: FeaturedPin[] = await api.get("/api/news/featured");
      const mapped: Article[] = (data || []).map((pin) => ({
        title: pin.title,
        description: pin.description || "",
        url: pin.url,
        urlToImage: pin.imageUrl || "",
        publishedAt: pin.pinnedAt,
        source: { name: pin.source || "Featured" },
      }));
      setArticles(mapped);
      setIsSearchResults(false);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const fetchSearch = async (q: string) => {
    setLoading(true);
    setError("");
    try {
      const data = await api.get(`/api/news?q=${encodeURIComponent(q)}`);
      setArticles(data || []);
      setIsSearchResults(true);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFeatured();
  }, []);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchInput.trim()) {
      fetchSearch(searchInput);
    }
  };

  const handleClear = () => {
    setSearchInput("");
    fetchFeatured();
  };

  return (
    <div>
      <form onSubmit={handleSearch} className="mb-8">
        <div className="flex gap-2 max-w-xl mx-auto">
          <input
            type="text"
            className="form-input flex-1 rounded border-gray-300 px-4 py-2"
            placeholder="Search news articles..."
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
          />
          <button
            type="submit"
            className="btn btn-primary px-6 py-2 rounded font-medium"
          >
            Search
          </button>
          {isSearchResults && (
            <button
              type="button"
              onClick={handleClear}
              className="px-4 py-2 rounded font-medium border border-gray-300 text-gray-600 hover:bg-gray-50"
            >
              Clear
            </button>
          )}
        </div>
      </form>

      {isSearchResults && (
        <p className="text-sm text-gray-500 mb-4">
          Search results for "{searchInput}" &mdash;{" "}
          <button onClick={handleClear} className="text-primary hover:underline">
            back to featured
          </button>
        </p>
      )}

      {loading && <p className="text-center text-gray-500">Loading articles...</p>}
      {error && <p className="text-center text-red-500">{error}</p>}

      <div className="row gy-5 gx-4">
        {articles.map((article, i) => (
          <div key={i} className={i === 0 ? "col-12" : "col-12 sm:col-6"}>
            {article.urlToImage && (
              <button
                onClick={() => onArticleClick(article)}
                className="rounded-lg block hover:text-primary overflow-hidden group w-full text-left"
              >
                <img
                  className="group-hover:scale-[1.03] transition duration-300 w-full rounded-lg"
                  src={article.urlToImage}
                  alt={article.title}
                  style={{ height: i === 0 ? 400 : 200, objectFit: "cover" }}
                  onError={(e) => { (e.target as HTMLImageElement).style.display = "none"; }}
                />
              </button>
            )}
            <ul className="mt-4 mb-2 flex flex-wrap items-center text-text text-sm">
              <li className="mr-4 font-medium text-primary">
                {article.source.name}
              </li>
              <li className="text-gray-500">
                {new Date(article.publishedAt).toLocaleDateString("en-US", {
                  month: "short",
                  day: "numeric",
                  year: "numeric",
                })}
              </li>
            </ul>
            <h3 className="mb-2">
              <button
                onClick={() => onArticleClick(article)}
                className="block hover:text-primary transition duration-300 text-left font-bold"
              >
                {article.title}
              </button>
            </h3>
            {article.description && (
              <p className="text-text line-clamp-2 text-sm">{article.description}</p>
            )}
          </div>
        ))}
      </div>

      {!loading && articles.length === 0 && !error && (
        <p className="text-center text-gray-500 mt-8">
          {isSearchResults ? "No articles found. Try a different search." : "No featured articles yet."}
        </p>
      )}
    </div>
  );
}

export type { Article };
