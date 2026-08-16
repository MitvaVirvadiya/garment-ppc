"use client";

import { roleLabel } from "@/lib/format";
import type { Plant, User } from "@/lib/types";
import { IconTimeline } from "@tabler/icons-react";
import Link from "next/link";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

export function AppShell({
  users,
  plants,
  user,
  plant,
  children,
}: {
  users: User[];
  plants: Plant[];
  user: string;
  plant: string;
  children: React.ReactNode;
}) {
  const path = usePathname();
  const router = useRouter();
  const params = useSearchParams();

  function setParam(key: string, value: string) {
    const next = new URLSearchParams(params.toString());
    next.set(key, value);
    if (key === "user") {
      const u = users.find((x) => x.slug === value);
      if (u?.role === "viewer" || u?.role === "plant_head") {
        const locked = plants.find((p) => p.id === u.plant_id);
        if (locked) next.set("plant", locked.code);
      }
    }
    router.push(`${path}?${next.toString()}`);
  }

  const current = users.find((u) => u.slug === user);

  return (
    <div className="page">
      <header className="loom-header">
        <div className="brand">
          <div className="brand-kicker">
            <IconTimeline size={14} aria-hidden="true" /> Loom control room
          </div>
          <h1 className="brand-title">Garment PPC</h1>
          <div className="brand-sub">
            See collisions this week across plants. Warp is the line. Time is the shuttle.
          </div>
        </div>
        <div className="header-tools">
          <nav className="nav-picks" aria-label="Primary">
            <Link href={`/?user=${user}&plant=${plant}`} aria-current={path === "/" ? "page" : undefined}>
              Warp
            </Link>
            <Link
              href={`/orders?user=${user}&plant=${plant}`}
              aria-current={path === "/orders" ? "page" : undefined}
            >
              Orders
            </Link>
          </nav>
          <div className="field">
            <label htmlFor="user-switch">Seat</label>
            <select
              id="user-switch"
              className="form-select"
              name="user"
              autoComplete="off"
              value={user}
              onChange={(e) => setParam("user", e.target.value)}
            >
              {users.map((u) => (
                <option key={u.slug} value={u.slug}>
                  {u.role === "viewer" ? "◎" : u.role === "merchandiser" ? "◇" : "▣"} {u.name} · {roleLabel(u.role)}
                </option>
              ))}
            </select>
          </div>
          <div className="field">
            <label htmlFor="plant-switch">Plant</label>
            <select
              id="plant-switch"
              className="form-select"
              name="plant"
              autoComplete="off"
              value={plant}
              disabled={current?.role === "viewer"}
              onChange={(e) => setParam("plant", e.target.value)}
            >
              {plants.map((p) => (
                <option key={p.code} value={p.code}>
                  {p.code} · {p.city}
                </option>
              ))}
            </select>
          </div>
        </div>
      </header>
      <main id="main" className="main-stage">
        {children}
      </main>
    </div>
  );
}
