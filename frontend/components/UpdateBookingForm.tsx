import React from 'react';
import DatePicker, { registerLocale } from 'react-datepicker';
import { formatAppointment } from '../utils/dateConversion';

interface Props {
    selectedBooking: { AppointmentStart: string } | null
    dateValue: Date | null
    setDateValue: (d: Date | null) => void
    durationValue: number | ''
    setDurationValue: (d: number) => void
    onSubmit: React.FormEventHandler<HTMLFormElement>
    errorMsg?: string
}

export default function UpdateBookingForm({
    selectedBooking,
    dateValue,
    setDateValue,
    durationValue,
    setDurationValue,
    onSubmit,
    errorMsg,
}: Props) {
    return (
        <form onSubmit={onSubmit}>
            {/* Heading */}
            <h2 className="text-3xl text-blue-50 font-bold text-center">
                Update Booking
            </h2>

            <div className="space-y-4">
                <label className="block mb-1 text-lg font-medium text-blue-50">
                    <span className="block">
                        From:{' '}
                        {selectedBooking
                            ? formatAppointment(selectedBooking.AppointmentStart)
                            : '—'}
                    </span>
                    <span>To:&nbsp;</span>
                    <DatePicker
                        selected={dateValue}
                        locale="en-US"
                        onChange={(date) => setDateValue(date)}
                        showTimeSelect
                        timeIntervals={30}
                        dateFormat="Pp"
                        className="bg-blue-200 text-blue-800 font-bold text-lg ml-6"
                    />
                </label>

                <label className="block">
                    <span className="block mb-1 text-lg text-blue-50 font-medium">
                        Duration in Minutes
                    </span>
                    <select
                        required
                        value={durationValue}
                        onChange={(e) => setDurationValue(Number(e.target.value))}
                        className="bg-blue-200 text-blue-800 font-bold text-lg ml-6"
                    >
                        <option value="" disabled>
                            Select duration
                        </option>
                        <option value={30}>30</option>
                        <option value={60}>60</option>
                    </select>
                </label>
            </div>

            {errorMsg && (
                <p className="text-red-300 text-sm text-center">{errorMsg}</p>
            )}

            <button
                type="submit"
                className="w-full text-blue-700 bg-blue-200 bg-opacity-20 hover:bg-opacity-30 py-2 
          rounded-lg font-semibold transition disabled:opacity-50 hover:bg-blue-400"
            >
                Submit Changes
            </button>
        </form>
    )
}
