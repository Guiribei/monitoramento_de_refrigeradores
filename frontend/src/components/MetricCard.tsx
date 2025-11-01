import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { LucideIcon } from "lucide-react";

interface MetricCardProps {
  title: string;
  value: string | number;
  unit?: string;
  icon: LucideIcon;
  trend?: "up" | "down" | "neutral";
}

export const MetricCard = ({ title, value, unit, icon: Icon }: MetricCardProps) => {
  return (
    <Card className="bg-gradient-card border-border hover:shadow-glow transition-all duration-300 hover:scale-105">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {title}
        </CardTitle>
        <Icon className="h-4 w-4 text-primary" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold text-foreground">
          {value}
          {unit && <span className="text-lg text-muted-foreground ml-1">{unit}</span>}
        </div>
      </CardContent>
    </Card>
  );
};
