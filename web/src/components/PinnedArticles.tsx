import React, { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { useAuth } from "./AuthContext";

interface Pin {
  userId: string;
  articleId: string;
  title: string;
  url: string;
  pinnedAt: string;
}

interface Props {
  onBack: () => void;
}

export default function PinnedArticles({ onBack }: Props) {
  const { isLoggedIn } = useAuth();
  const [pins, setPins] = useState<Pin[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchPins = async () => {
    setLoading(true);
    try {
      const data = await api.get("/api/pins");
      setPins(data || []);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isLoggedIn) fetchPins();
  }, [isLoggedIn]);

  const handleUnpin = async (articleId: string) => {
    try {
      await api.delete(`/api/pins?articleId=${encodeURIComponent(articleId)}`);
      setPins(pins.filter((p) => p.articleId !== articleId));
    } catch (err: any) {
      setError(err.message);
    }
  };

  if (!isLoggedIn) {
    return (
      <div className="text-center mt-12">
        <p className="text-gray-500">Sign in to view your pinned articles.</p>
      </div>
    );
  }

  return (
    <div>
      <button
        onClick={onBack}
        className="mb-6 text-primary hover:underline font-medium"
      >
        &larr; Back to feed
      </button>

      <h2 className="text-2xl font-bold mb-6">My Pinned Articles</h2>

      {loading && <p className="text-gray-500">Loading pins...</p>}
      {error && <p className="text-red-500">{error}</p>}

      {!loading && pins.length === 0 && (
        <p className="text-gray-500">No pinned articles yet. Browse the feed and pin articles you like.</p>
      )}

      <div className="space-y-4">
        {pins.map((pin) => (
          <div
            key={pin.articleId}
            className="bg-white rounded-lg shadow-sm border p-4 flex items-center justify-between gap-4"
          >
            <div className="flex-1 min-w-0">
              <h3 className="font-bold truncate">{pin.title}</h3>
              <p className="text-sm text-gray-500">
                Pinned {new Date(pin.pinnedAt).toLocaleDateString("en-US", {
                  month: "short",
                  day: "numeric",
                  year: "numeric",
                })}
              </p>
            </div>
            <div className="flex gap-2 shrink-0">
              <a
                href={pin.url}
                target="_blank"
                rel="noopener noreferrer"
                className="btn btn-primary px-4 py-1 rounded text-sm font-medium"
              >
                Read
              </a>
              <button
                onClick={() => handleUnpin(pin.articleId)}
                className="px-4 py-1 rounded text-sm font-medium border border-red-400 text-red-500 hover:bg-red-50 transition"
              >
                Unpin
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
