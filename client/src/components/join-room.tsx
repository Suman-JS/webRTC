import { type FormEvent, memo, useState } from "react";

import type { JoinRoomProps } from "@/types";

export const JoinRoom = memo(({ onJoin }: JoinRoomProps) => {
  const [roomId, setRoomId] = useState<string>("");
  const [username, setUsername] = useState<string>("");

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (roomId.trim() && username.trim()) {
      onJoin(roomId, username);
    }
  };


  return (
    <div className="max-w-300 mx-auto bg-purple-400 p-8 rounded-lg shadow-md">
      <h2>
        Join a room
      </h2>

      <form onSubmit={handleSubmit}>
        <div className="mb-6">
          <label htmlFor="username">Your name</label>
          <input
            type="text"
            id="username"
            value={username}
            placeholder="Enter your name"
            required
            onChange={(e) => setUsername(e.target.value)}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
          />
        </div>

        <div className="mb-6">
          <label htmlFor="roomId">Room ID</label>
          <input
            type="text"
            id="roomId"
            placeholder="Enter room ID"
            required
            value={roomId}
            onChange={(e) => setRoomId(e.target.value)}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
          />
        </div>
        <button
          type="submit"
          className="w-full rounded-md bg-purple-500 px-4 py-2 text-white shadow-sm hover:bg-purple-600 focus:outline-none focus:ring-2 focus:ring-purple-400 focus:ring-offset-2 focus:ring-offset-gray-800"
        >
          Join
        </button>
      </form>
    </div>
  );

});
