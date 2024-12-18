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
  active_users: number;
  users: string[];
}

export interface Message {
  user: string;
  type: "JOIN" | "LEAVE" | "CHAT";
  content: string;
}
