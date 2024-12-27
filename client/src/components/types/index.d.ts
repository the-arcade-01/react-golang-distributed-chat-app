export interface AuthResponse {
  message: string;
  data: {
    user_id: number;
    username: string;
    token: string;
  };
}

export interface RoomDetails {
  room_id: string;
  room_name: string;
  admin: string;
}

export interface Message {
  user: string;
  type: "JOIN" | "LEAVE" | "CHAT";
  content: string;
}
