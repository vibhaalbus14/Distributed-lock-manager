import { useState } from "react";

// 🚀 Imports our scoped styling layouts
import styles from "./network.module.css";

export default function Network({ logMsg }) {
  const [minLatency, setMinLatency] = useState(0);
  const [maxLatency, setMaxLatency] = useState(0);
  const [droprate, setDroprate] = useState(0);

  function minLatencyChange(e) {
    if(parseInt(e.target.value)>maxLatency){
      alert("Minimum Latency cannot be higher than Maximum Latency, please update the value")
      return
    }
    setMinLatency(parseInt(e.target.value));
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

  // Adjusted onChange references below to execute closures passing event (e) targets
  function maxLatencyChange(e) {
    if(parseInt(e.target.value)<minLatency){
      alert("Maximum Latency cannot be lower than Minimum Latency, please update the value")
      return
    }
    setMaxLatency(parseInt(e.target.value));
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
          <label className={styles.minLatency}>Min Latency: <span>{minLatency}ms</span></label>
          <div className={styles.HiddenBoxMin}> <div className={styles.tooltip_title}>
    <span>How is latency calculated in network</span>
    <span className={styles.info_icon}>?</span>
  </div>
  <div className={styles.tooltip_body}>
   Calculates a random delay time for each packet between your minimum and maximum settings. It starts with your minimum delay as the base, then adds a random slice of time up to your maximum limit. This mimics real-world internet lag, which constantly fluctuates instead of staying perfectly flat.
  </div></div>
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
          <label className={styles.maxLatency}>Max Latency: <span>{maxLatency}ms</span></label>
          <div className={styles.HiddenBoxMax}>
            <div className={styles.tooltip_title}>
    <span>How is latency calculated in network</span>
    <span className={styles.tooltip_icon}>?</span>
  </div>
  
  {/* Part 2: The Core Description Body */}
  <div className={styles.tooltip_body}>
   Calculates a random delay time for each packet between your minimum and maximum settings. It starts with your minimum delay as the base, then adds a random slice of time up to your maximum limit. This mimics real-world internet lag, which constantly fluctuates instead of staying perfectly flat.
  </div></div>
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
          <label className={styles.droprate}>Drop Rate: <span>{(parseFloat(droprate) * 100).toFixed(0)}%</span></label>
          <div className={styles.HiddenBoxDroprate}>
  {/* Part 1: The Title Heading Row */}
  <div className={styles.tooltip_title}>
    <span>How is droprate calculated in network</span>
    <span className={styles.tooltip_icon}>?</span>
  </div>
  
  {/* Part 2: The Core Description Body */}
  <div className={styles.tooltip_body}>
    Generates a random percentage for each incoming packet. If that number falls below your chosen drop rate threshold, 
    the packet is instantly discarded to simulate network loss. This tests how effectively the cluster recovers from 
    missing messages and maintains consensus under unstable conditions.
  </div>
</div>
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