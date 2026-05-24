import { useEffect, useReducer } from 'react'; 
// Fixed file casing imports to match your project layouts
import NodeGroup from './components/nodeGroup';     
import Network from './components/network';
import LockManager from './components/lockManager';
import appStyles from './app.module.css';

const initialData = { fencingToken: 0, currentHolder: "", nodeStatus: {}, opTime: 0 ,logMsg:{}};
function reducer(state, action) {
  switch (action.type) {
    case "NODE_DELTA":
     
      return {
        ...state,
        nodeStatus: {
          ...state.nodeStatus,
          [action.payload.NodeId]: action.payload.Status
        }
      };

    case "OP":
      
      return {
        ...state,
        currentHolder: action.payload.CurrentHolder,
        fencingToken: action.payload.FencingToken,
        // Parse your string-formatted duration safely into a Base-10 integer
        opTime: parseInt(action.payload.OpTime, 10) || 0 
      };

    case "LOG":
      
      return {
        ...state,
        logMsg:action.payload//this contains event : and message :
      };
    case "CHAN_FREE":
      
      return {  
        ...state,
            currentHolder:"",opTime:0,logMsg:{}
      };
    case "EXIT_APP":
      
      return initialData


    default:
      console.warn(`[WS] Unhandled payload channel type received: ${action.Type}`);
  break;
  }
}

export default function App() {
  
  const [clusterData, dispatcher] = useReducer(reducer, initialData);

  // Multi-Channel Live WebSocket Data Listener
  useEffect(() => {
    const ws = new WebSocket(`${import.meta.env.VITE_API_WS_URL}/network/stream`);
    console.log("ws connection established successfully from react client")
    ws.onmessage = (event) => {
      
      const envelope = JSON.parse(event.data);
      console.log(envelope)
      // Traffic Cop: Route data frames depending on payload text type tags
      switch (envelope.Type) {
        case "NODE_DELTA":
          dispatcher({ type: "NODE_DELTA", payload: envelope.Payload });
          break;
        case "OP":
          dispatcher({ type: "OP", payload: envelope.Payload });
          break;
        case "LOG":
          dispatcher({ type: "LOG", payload: envelope.Payload });
          break;
        case "CHAN_FREE":
          dispatcher({ type: "CHAN_FREE" });//payload not necessary here
          break;
        case "EXIT_APP":
          dispatcher({ type: "EXIT_APP" });//payload not necessary here
          break;
        default:
          // Raw string logs ("LOG") bypass this and are captured inside network.jsx
          throw new Error("Unsupported action")
      }
    };

    ws.onerror = (err) => console.error("Websocket connection failed", err);
    return () => ws.close();
  }, []); 

  const handleExitServer = async () => {
    const confirmShutdown = window.confirm("Are you sure you want to completely restart the Go Lock manager,this will delete all existing nodes?");
    if (!confirmShutdown) return;

    try {
      const res = await fetch(`${import.meta.env.VITE_API_HTTP_URL}/exit`, {
        method: "GET", // Automatically matches your preferred endpoint router configuration
      });
      
      if (!res.ok) {
       console.error("Failed to execute server shutdown command sequence:",res.ok);
      }
    } catch (err) {
      console.error("Could not execute exit", err);
    }
  };
  

  // Destructure state variables to pass into UI layouts
  const { nodeStatus, currentHolder, fencingToken, opTime,logMsg } = clusterData;

  return (
    <div style={{ padding: "20px", background: "#111", minHeight: "100vh", color: "#fff" }}>
      <h2><center>Distributed Lock Manager Dashboard</center></h2>
      <div className={appStyles.exitActionContainer}>
        <button onClick={handleExitServer} className={appStyles.exitBtn}>
          🛑Restart Cluster Engine
        </button>
      </div>
      
      {/* Visual node status grid matrix layout */}
      <NodeGroup nodes={nodeStatus} />
      
      {/* Latency and drop rate adjustment slider controllers */}
      <Network logMsg={logMsg}/>
      
      {/* Central coordinator readouts & active task leasing countdown block */}
      <LockManager 
        fencingToken={fencingToken} 
        currentHolder={currentHolder} 
        opTime={opTime} 
      />
    </div>
  );
}