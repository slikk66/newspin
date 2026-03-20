import React, { useState } from "react";
import { useAuth } from "./AuthContext";
import { api } from "@/lib/api";
import type { Article } from "./NewsFeed";

interface Props {
  article: Article;
  onBack: () => void;
}

export default function ArticleView({ article, onBack }: Props) {
  const { isLoggedIn } = useAuth();
  const [pinned, setPinned] = useState(false);
  const [pinLoading, setPinLoading] = useState(false);
  const [message, setMessage] = useState("");

  const articleId = btoa(article.url).slice(0, 64);

  const handlePin = async () => {
    setPinLoading(true);
    setMessage("");
    try {
      await api.post("/api/pins", {
        articleId,
        title: article.title,
        url: article.url,
        description: article.description,
        imageUrl: article.urlToImage,
        source: article.source.name,
      });
      setPinned(true);
      setMessage("Pinned!");
    } catch (err: any) {
      setMessage(err.message);
    } finally {
      setPinLoading(false);
    }
  };

  const handleUnpin = async () => {
    setPinLoading(true);
    try {
      await api.delete(`/api/pins?articleId=${encodeURIComponent(articleId)}`);
      setPinned(false);
      setMessage("Unpinned");
    } catch (err: any) {
      setMessage(err.message);
    } finally {
      setPinLoading(false);
    }
  };

  return (
    <div className="max-w-3xl mx-auto">
      <button
        onClick={onBack}
        className="mb-6 text-primary hover:underline font-medium"
      >
        &larr; Back to feed
      </button>

      {article.urlToImage && (
        <img
          src={article.urlToImage}
          alt={article.title}
          className="w-full rounded-lg mb-6"
          style={{ maxHeight: 450, objectFit: "cover" }}
          onError={(e) => { (e.target as HTMLImageElement).style.display = "none"; }}
        />
      )}

      <div className="flex items-center gap-3 mb-4 text-sm">
        <span className="font-medium text-primary">{article.source.name}</span>
        <span className="text-gray-500">
          {new Date(article.publishedAt).toLocaleDateString("en-US", {
            month: "long",
            day: "numeric",
            year: "numeric",
          })}
        </span>
      </div>

      <h1 className="text-3xl font-bold mb-4">{article.title}</h1>
      <p className="text-lg text-text mb-8">{article.description}</p>

      <div className="flex gap-3 items-center flex-wrap">
        <a
          href={article.url}
          target="_blank"
          rel="noopener noreferrer"
          className="btn btn-primary px-6 py-2 rounded font-medium inline-block"
        >
          Read Source Article &rarr;
        </a>

        {isLoggedIn && (
          <button
            onClick={pinned ? handleUnpin : handlePin}
            disabled={pinLoading}
            className={`px-6 py-2 rounded font-medium border-2 transition ${
              pinned
                ? "border-red-400 text-red-500 hover:bg-red-50"
                : "border-primary text-primary hover:bg-primary hover:text-white"
            }`}
          >
            {pinLoading ? "..." : pinned ? "Unpin" : "Pin Article"}
          </button>
        )}

        {message && <span className="text-sm text-gray-600">{message}</span>}
      </div>

      {!isLoggedIn && (
        <p className="mt-4 text-sm text-gray-500">
          Sign in to pin articles to your collection.
        </p>
      )}
    </div>
  );
}
