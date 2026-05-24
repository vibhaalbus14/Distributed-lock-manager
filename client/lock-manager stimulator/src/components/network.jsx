import { useState } from "react";

// 🚀 Imports our scoped styling layouts
import styles from "./network.module.css";

export default function Network({ logMsg }) {
  const [minLatency, setMinLatency] = useState(0);
  const [maxLatency, setMaxLatency] = useState(0);
  const [droprate, setDroprate] = useState(0);

  function minLatencyChange(e) {
    setMinLatency(e.target.value);
    async function callLatency() {
      try {
        const res = await fetch(`${import.meta.env.VITE_API_HTTP_URL}/network/latency/minLatency=${minLatency}&maxLatency=${maxLatency}`, {
          method: "PUT",
        });
        if (!res.ok) {
          throw new Error(`HTTP error! status: ${res.status}`);
        }
      } catch (err) {
        console.error("Failed to connect to Go backend:", err);
      }
    }
    callLatency();
  }

  // Adjusted onChange references below to execute closures passing event (e) targets
  function maxLatencyChange(e) {
    setMaxLatency(e.target.value);
    async function callLatency() {
      try {
        const res = await fetch(`${import.meta.env.VITE_API_HTTP_URL}/network/latency?minLatency=${minLatency}&maxLatency=${maxLatency}`, {
          method: "PUT",
        });
        if (!res.ok) {
          throw new Error(`HTTP error! status: ${res.status}`);
        }
      } catch (err) {
        console.error("Failed to connect to Go backend:", err);
      }
    }
    callLatency();
  }

  function droprateChange(e) {
    setDroprate(e.target.value);
    async function callDroprate() {
      try {
        const res = await fetch(`${import.meta.env.VITE_API_HTTP_URL}/network/droprate?rate=${droprate}`, {
          method: "PUT",
        });
        if (!res.ok) {
          throw new Error(`HTTP error! status: ${res.status}`);
        }
      } catch (err) {
        console.error("Failed to connect to Go backend:", err);
      }
    }
    callDroprate();
  }

  // Simple look-up selector to determine the message box border colors
  let alertThemeClass = "";
  if (logMsg.Event === "PACKET_DROPPED") {
    alertThemeClass = styles.dropAlert;
  } else if (logMsg.Event === "PACKET_DELAYED") {
    alertThemeClass = styles.delayAlert;
  }

  return (
    <div className={styles.networkCard}>
      <h1>Network Simulation Engine</h1>
      
     
      <div className={styles.sliderGrid}>
        
        <div className={styles.controlGroup}>
          <label>Min Latency: <span>{minLatency}ms</span></label>
          <input 
            type="range" 
            min="0" 
            max="50" 
            step="2" 
            value={minLatency} 
            className={styles.sliderInput}
            onChange={minLatencyChange} 
          />
        </div>

        <div className={styles.controlGroup}>
          <label>Max Latency: <span>{maxLatency}ms</span></label>
          <input 
            type="range" 
            min="0" 
            max="50" 
            step="2" 
            value={maxLatency} 
            className={styles.sliderInput}
            onChange={maxLatencyChange} 
          />
        </div>

        <div className={styles.controlGroup}>
          <label>Drop Rate: <span>{(parseFloat(droprate) * 100).toFixed(0)}%</span></label>
          <input 
            type="range" 
            min="0.0" 
            max="1.0" 
            step="0.1" 
            value={droprate} 
            className={styles.sliderInput}
            onChange={droprateChange} 
          />
        </div>

      </div>

     
      <div className={`${styles.msgBox} ${alertThemeClass}`}>
        {!logMsg.Event ? (
          <span className={styles.normalState}>🟢 Network Tunnel Operating Normally</span>
        ) : (
          <>
            <h3 className={styles.chaosHeader}>⚠️ {logMsg.Event}</h3>
            <p className={styles.chaosBody}>{logMsg.Message}</p>
          </>
        )}
      </div>
    </div>
  );
}