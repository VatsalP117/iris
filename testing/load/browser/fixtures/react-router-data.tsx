import React, { useEffect } from "react";
import { createRoot } from "react-dom/client";
import {
    Link,
    Outlet,
    RouterProvider,
    createBrowserRouter,
    redirect,
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

function Layout() {
    const navigate = useNavigate();

    useEffect(() => {
        const iris = new Iris({
            host: location.origin,
            siteId: "browser-lab-react-router-data",
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
            <Link id="account-link" to="/framework/react-data/account">
                Account
            </Link>
            <Link id="redirect-link" to="/framework/react-data/legacy">
                Legacy redirect
            </Link>
            <button
                id="replace-button"
                onClick={() =>
                    navigate("/framework/react-data/settings", {
                        replace: true,
                    })
                }
            >
                Replace with settings
            </button>
            <Outlet />
        </main>
    );
}

const router = createBrowserRouter([
    {
        path: "/framework/react-data",
        element: <Layout />,
        children: [
            { index: true, element: <p>Home</p> },
            { path: "account", element: <p>Account</p> },
            { path: "settings", element: <p>Settings</p> },
            { path: "legacy", loader: () => redirect("/framework/react-data/account") },
        ],
    },
]);

createRoot(document.getElementById("root")!).render(
    <RouterProvider router={router} />,
);
