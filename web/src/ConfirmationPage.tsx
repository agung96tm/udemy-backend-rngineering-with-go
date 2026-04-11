import {API_URL} from "./App.tsx";
import { useParams, useNavigate } from 'react-router-dom';

export const ConfirmationPage = () => {
    const { token = "" } = useParams();
    const redirect = useNavigate();

    const confirmHandler = async () => {
        const resp = await fetch(`${API_URL}/users/activate/${token}`, {
            method: "PUT",
        });

        if (resp.ok) {
            redirect("/");
        } else {
            alert("Failed to confirm token");
        }
    }

    return (
        <div>
            <h1>Confirmation</h1>
            <div>
                <button onClick={confirmHandler}>Click to Confirm</button>
            </div>
        </div>
    )
}