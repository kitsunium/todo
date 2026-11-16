import { QueryClientProvider } from "@tanstack/react-query";
import { MotionConfig } from "motion/react";
import { useState } from "react";
import { BrowserRouter } from "react-router";
import { errorMessage, isUnauthenticated } from "../api/errors";
import { makeQueryClient } from "../api/queries";
import { TooltipProvider } from "../components/ui/tooltip";
import { toast, Toaster } from "../components/ui/toast";
import { AppRoutes } from "./routes";

export function App() {
  const [client] = useState(() =>
    makeQueryClient((err) => {
      if (!isUnauthenticated(err)) toast.error(errorMessage(err));
    }),
  );
  return (
    <QueryClientProvider client={client}>
      <MotionConfig reducedMotion="user">
        <TooltipProvider delayDuration={500} skipDelayDuration={200}>
          <BrowserRouter>
            <AppRoutes />
          </BrowserRouter>
          <Toaster />
        </TooltipProvider>
      </MotionConfig>
    </QueryClientProvider>
  );
}
