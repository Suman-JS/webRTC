import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import type { MessageResponse } from "@/types";


export const Route = createFileRoute("/")({
  component: App,
});

function App() {
  const [inRoom, setInRoom] = useState(false);
  const [roomId, setRoomId] = useState("");
  const [username, setUsername] = useState("");
  const [peers, setPeers] = useState<{
    peerId: string;
    peerUsername: string;
  }>({ peerId: "", peerUsername: "" });
  const [clientId, setClientId] = useState<string | null>(null);

  const socketRef = useRef<WebSocket | null>(null);
  const localStreamRef = useRef<MediaStream | null>(null);
  const peerConnectionRef = useRef({});

  useEffect(() => {
    socketRef.current = new WebSocket("ws://localhost:8080/ws");

    socketRef.current.onopen = () => {
      console.log("WebSocket connection established");
    };

    socketRef.current.onmessage = (event) => {
      const message = JSON.parse(event.data);
      handleSignalingMessage(message);
    };

    socketRef.current.onclose = () => {
      console.log("WebSocket connection closed");
    };

    return () => {
      socketRef.current?.close();
      if (localStreamRef.current) {
        localStreamRef.current?.getTracks().forEach(track => track.stop());
      }
      Object.values(peerConnectionRef.current).forEach(pc => pc.close());
    };
  }, []);

  const handleSignalingMessage = async (message: MessageResponse) => {
    console.log("Received signaling message:", message);

    switch (message.type) {
      case "init":
        setClientId(message.data.clientId);
        break;

      case "room-joined":
        setInRoom(true);
        const roomPeers = message.data.peers;
        await setupLocalMedia();
        roomPeers.forEach(peer => {
          createPeerConnection(peer.id, peer.username, true);
        });
        break;

      case "new-peer":
        const newPeerId = message.sender;
        const newPeerUsername = message.data.username;
        await setupLocalMedia();
        createPeerConnection(newPeerId, newPeerUsername, false);
        break;

      case "peer-left":
        const peerId = message.sender;

        if (peerConnectionRef.current[peerId]) {
          peerConnectionRef.current[peerId].close();
          delete peerConnectionRef.current[peerId];

          setPeers(prevPeers => {
            const newPeers = { ...prevPeers };
            delete newPeers[peerId];
            return newPeers;
          });
          break;
        }
    }
  };


  return (
    <div>
      <p className="bg-amber-500 mx-auto m-0 p-4">
        Welcome to webRTC
      </p>
    </div>
  );
}
