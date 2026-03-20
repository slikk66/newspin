import React, { useState } from "react";
import { AuthProvider, useAuth } from "./AuthContext";
import NewsFeed from "./NewsFeed";
import ArticleView from "./ArticleView";
import PinnedArticles from "./PinnedArticles";
import LoginPage from "./LoginPage";
import type { Article } from "./NewsFeed";

type View = "feed" | "article" | "pins" | "login";

function AppInner() {
  const { isLoggedIn, username, logout } = useAuth();
  const [view, setView] = useState<View>("feed");
  const [selectedArticle, setSelectedArticle] = useState<Article | null>(null);

  const handleArticleClick = (article: Article) => {
    setSelectedArticle(article);
    setView("article");
    window.scrollTo(0, 0);
  };

  const handleBack = () => {
    setView("feed");
    setSelectedArticle(null);
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-8 pb-4 border-b">
        <nav className="flex gap-4">
          <button
            onClick={() => { setView("feed"); setSelectedArticle(null); }}
            className={`font-medium transition ${view === "feed" ? "text-primary" : "text-gray-500 hover:text-dark"}`}
          >
            Feed
          </button>
          {isLoggedIn && (
            <button
              onClick={() => setView("pins")}
              className={`font-medium transition ${view === "pins" ? "text-primary" : "text-gray-500 hover:text-dark"}`}
            >
              My Pins
            </button>
          )}
        </nav>
        <div className="flex items-center gap-3">
          {isLoggedIn ? (
            <>
              <span className="text-sm text-gray-600">Hi, {username}</span>
              <button
                onClick={logout}
                className="text-sm text-gray-500 hover:text-red-500 transition"
              >
                Logout
              </button>
            </>
          ) : (
            <button
              onClick={() => setView("login")}
              className="btn btn-primary px-4 py-1 rounded text-sm font-medium"
            >
              Sign In
            </button>
          )}
        </div>
      </div>

      {view === "feed" && <NewsFeed onArticleClick={handleArticleClick} />}
      {view === "article" && selectedArticle && (
        <ArticleView article={selectedArticle} onBack={handleBack} />
      )}
      {view === "pins" && <PinnedArticles onBack={handleBack} />}
      {view === "login" && <LoginPage onSuccess={() => setView("feed")} />}
    </div>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <AppInner />
    </AuthProvider>
  );
}
