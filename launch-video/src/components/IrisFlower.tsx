import type { CSSProperties } from "react";

interface Props {
  color?: string;
  size?: number;
  style?: CSSProperties;
}

export const IrisFlower: React.FC<Props> = ({ color = "currentColor", size = 96, style }) => {
  const petalWidth = size * 0.2;
  const petalHeight = size * 0.46;

  return (
    <div style={{ height: size, position: "relative", width: size, ...style }}>
      {Array.from({ length: 6 }, (_, index) => (
        <div
          key={index}
          style={{
            backgroundColor: color,
            borderRadius: "999px 999px 160px 160px",
            height: petalHeight,
            left: size * 0.4,
            position: "absolute",
            rotate: `${index * 60}deg`,
            top: size * 0.04,
            transformOrigin: `${petalWidth / 2}px ${petalHeight}px`,
            width: petalWidth,
          }}
        />
      ))}
    </div>
  );
};
