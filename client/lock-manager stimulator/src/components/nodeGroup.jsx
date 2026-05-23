import Node from './node';

import styles from './nodeGroup.module.css';

export default function NodeGroup({ nodes }) {

  async function createNode() {
    try {
      const res = await fetch("http://localhost:8080/nodes", {
        method: "GET", 
      });

      if (!res.ok) {
        throw new Error("Unable to create new node");
      }
    } catch (err) {
      console.error("Creation failed:", err);
    }
  }

  return (
    <div className={styles.groupContainer}>
      
      <h1 className={styles.centeredHeading}>NODES</h1>
      
     
      <div className={styles.flexGrid}>
        {Object.keys(nodes).map((id) => (
          <Node
            key={id}
            id={id}
            status={nodes[id]}
          />
        ))}
      </div>

     
      <div className={styles.actionFooter}>
        <button onClick={createNode} className={styles.createBtn}>
          Create new node
        </button>
      </div>
    </div>
  );
}