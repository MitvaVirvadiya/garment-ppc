export function fmtDay(iso: string, timeZone = "Asia/Kolkata") {
  return new Intl.DateTimeFormat("en-IN", {
    weekday: "short",
    day: "2-digit",
    month: "short",
    timeZone,
  }).format(new Date(iso));
}

export function fmtTime(iso: string, timeZone = "Asia/Kolkata") {
  return new Intl.DateTimeFormat("en-IN", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
    timeZone,
  }).format(new Date(iso));
}

export function fmtWhen(iso: string, timeZone = "Asia/Kolkata") {
  return new Intl.DateTimeFormat("en-IN", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
    timeZone,
  }).format(new Date(iso));
}

export function fmtQty(n: number) {
  return new Intl.NumberFormat("en-IN").format(n);
}

export function fmtPct(n: number) {
  return `${n.toFixed(n >= 10 ? 0 : 1)}%`;
}

export function hoursBetween(a: string, b: string) {
  return (new Date(b).getTime() - new Date(a).getTime()) / 36e5;
}

export function addHours(iso: string, hours: number) {
  return new Date(new Date(iso).getTime() + hours * 36e5).toISOString();
}

export function roleLabel(role: string) {
  if (role === "plant_head") return "Plant head";
  if (role === "merchandiser") return "Merchandiser";
  if (role === "viewer") return "Viewer";
  return role;
}
