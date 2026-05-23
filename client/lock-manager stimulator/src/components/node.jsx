import React from 'react';
import nodeIcon from '../assets/node.svg';

// 🚀 Local module tracking styles
import styles from './Node.module.css';

const STATUS_MAP = {
  0: "Idle",
  1: "Requesting",
  2: "Holding",
  3: "Crashed"
};

export function Node({ status, id }) {
  
  const handleRequestLock = async () => {
    try {
      await fetch("http://localhost:8080/nodes/request", {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: id }) 
      });
    } catch (err) {
      console.error("Failed to request lock:", err);
    }
  };

  const handleKillNode = async () => {
    try {
      await fetch("http://localhost:8080/nodes/kill", {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: id }) 
      });
    } catch (err) {
      console.error("Failed to kill node:", err);
    }
  };

  const handleResetNode = async () => {
    try {
      await fetch("http://localhost:8080/nodes/restart", {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: id }) 
      });
    } catch (err) {
      console.error("Failed to reset node:", err);
    }
  };

  // Determine button layout settings based on status
  let buttonText = "";
  let handleButtonClick = null;
  status = Number(status);

  if (status === 0) {
    buttonText = "Request";
    handleButtonClick = handleRequestLock;
  } else if (status === 2) {
    buttonText = "Kill";
    handleButtonClick = handleKillNode;
  } else if (status === 3) {
    buttonText = "Reset";
    handleButtonClick = handleResetNode;
  }

  const statusString = STATUS_MAP[status] || "Idle";

  return (
   
    <div className={`${styles.nodeCard} ${styles[statusString]}`}>
      <img src={nodeIcon} alt="node-img" width="50" />
      
      <h3>Node ID: {id.slice(0, 5)}</h3>
      <p>Status: <strong>{statusString}</strong></p>
      
      {buttonText && (
        <button className={styles.actionButton} onClick={handleButtonClick}>
          {buttonText}
        </button>
      )}
    </div>
  );
}

export default React.memo(Node);