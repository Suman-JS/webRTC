type JoinRoomProps = {
  onJoin: (roomId: string, username: string) => void;
}

type MessageResponse = {
  type: "init" | "room-joined" | "new-peer" | "peer-left" | "offer" | "answer" | "ice-candidate";
  sender: string;
  data: {
    username: string;
    peers: {
      id: string;
      username: string;
    }[];
    clientId: string;


  }
}

export type {
  JoinRoomProps,
  MessageResponse,
};