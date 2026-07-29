import React, { useEffect } from "react";
import { createRoot } from "react-dom/client";
import {
    BrowserRouter,
    Link,
    Route,
    Routes,
    useNavigate,
} from "react-router-dom";

import { Iris } from "../../../../web/src/index";

declare global {
    interface Window {
        __fixtureReady?: boolean;
        __irisReady?: boolean;
        __fixtureIris?: Iris;
    }
}

function App() {
    const navigate = useNavigate();

    useEffect(() => {
        const iris = new Iris({
            host: location.origin,
            siteId: "browser-lab-react-router-declarative",
            autocapture: { pageviews: true },
        });
        iris.start();
        window.__fixtureIris = iris;
        window.__fixtureReady = true;
        window.__irisReady = true;
        return () => iris.stop();
    }, []);

    return (
        <main>
            <nav>
                <Link id="pricing-link" to="/framework/react-declarative/pricing">
                    Pricing
                </Link>
                <Link id="same-link" to={location.pathname}>
                    Current route
                </Link>
                <button
                    id="docs-button"
                    onClick={() =>
                        navigate("/framework/react-declarative/docs")
                    }
                >
                    Docs
                </button>
            </nav>
            <Routes>
                <Route path="*" element={<p id="route">{location.pathname}</p>} />
            </Routes>
        </main>
    );
}

createRoot(document.getElementById("root")!).render(
    <BrowserRouter>
        <App />
    </BrowserRouter>,
);
