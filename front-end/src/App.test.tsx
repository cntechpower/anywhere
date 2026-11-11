import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { vi } from "vitest";
import App from "./App";

// Keep the shell test focused on rendering: mocking the two network helpers
// avoids pulling axios into the jsdom environment.
vi.mock("./api/client", () => ({ apiPostForm: vi.fn() }));
vi.mock("./apis/apis", () => ({
  default: { userLogoutApi: "/user_logout" },
}));

test("renders the console navigation", () => {
  render(
    <MemoryRouter initialEntries={["/home"]}>
      <App>
        <div>content</div>
      </App>
    </MemoryRouter>
  );

  expect(screen.getByText("配置列表")).toBeInTheDocument();
  expect(screen.getByText("新增配置")).toBeInTheDocument();
  expect(screen.getAllByText("概览").length).toBeGreaterThan(0);
});
