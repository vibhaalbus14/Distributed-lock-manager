import styles from './lockManager.module.css';

export default function LockManager({ fencingToken, currentHolder, opTime }) {
  return (
    <div className={styles.managerCard}>
      <h2>Central Lock Coordinator</h2>
      
     
      <div className={styles.verticalStack}>
        
        <div className={styles.metricLine}>
          <strong>Fencing Token:</strong>
          <span className={styles.yellowHighlight}>{fencingToken}</span>
        </div>

        <div className={styles.metricLine}>
          <strong>Current Holder:</strong>
          {currentHolder ? (
            /* Safe slicing guard prevents crashes on unassigned fields */
            <span className={styles.yellowHighlight}>{currentHolder.slice(0, 5)}</span>
          ) : (
            <span className={styles.idleValue}>None (Lock Available)</span>
          )}
        </div>

        <div className={styles.metricLine}>
          <strong>Lock being held for:</strong>
          <span className={styles.yellowHighlight}>{opTime} seconds</span>
        </div>

      </div>

    
      {currentHolder && opTime > 0? (
        <div className={styles.countdownBox}>
          ⏱️ Node execution active: performing distributed task operations for {opTime}s before releasing lock...
        </div>
      ):<div className={styles.countdownBox}>Lock is Available, nodes are free to request</div>}
    </div>
  );
}